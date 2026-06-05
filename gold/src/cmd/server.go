package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Sn0wye/snowflake/gold/pkg/config"
	jwtpkg "github.com/Sn0wye/snowflake/gold/pkg/jwt"
	"github.com/Sn0wye/snowflake/gold/pkg/logger"
	"github.com/Sn0wye/snowflake/gold/pkg/messaging"
	"github.com/Sn0wye/snowflake/gold/pkg/middleware"
	"github.com/Sn0wye/snowflake/gold/src/repository"
	"github.com/Sn0wye/snowflake/gold/src/service"
	"github.com/Sn0wye/snowflake/gold/pkg/validator"
	"github.com/Sn0wye/snowflake/gold/src/db"
	"github.com/Sn0wye/snowflake/gold/src/migration"
	"github.com/Sn0wye/snowflake/gold/src/models"
	"github.com/Sn0wye/snowflake/gold/src/reconciliation"
	"github.com/Sn0wye/snowflake/gold/src/routes"

	"github.com/google/uuid"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

//	@title			Snowflake API Reference
//	@version		1.0.0
//	@description	The Snowflake API is organized around REST. This API has predictable resource-oriented URLs, accepts JSON-encoded request bodies, returns JSON-encoded responses, and uses standard HTTP response codes, authentication, and verbs.

//	@contact.name	GitHub
//	@contact.url	https://github.com/Sn0wye/snowflake/issues

//	@license.name	GNU General Public License v3.0
//	@license.url	https://www.gnu.org/licenses/gpl-3.0

// @host		snowflake.snowye.dev
// @BasePath	/

// @Schemes	https
func main() {
	flag.Parse()

	conf := config.GetConfig()
	logger := logger.NewLog(conf)

	migrateDB(logger)

	// Start RabbitMQ and defer its closure
	rmq := startRabbitMQ(conf, logger)
	defer rmq.Close()

	// Channel to handle shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start reconciliation job (hourly)
	reconcileJob := reconciliation.NewJob(db.GetDB(), logger)
	reconcileJob.Start(time.Hour)

	// Start servers
	go startHTTPServer(conf, logger, rmq, reconcileJob)
	go startAccountCreationConsumer(rmq, logger)

	<-quit // Wait for shutdown signal

	log.Println("Shutting down the servers...")
	log.Println("All servers stopped gracefully")
}

func migrateDB(logger *logger.Logger) {
	db := db.GetDB()

	migrate := migration.NewMigrate(db, logger)
	migrate.Run()
}

func startRabbitMQ(conf *viper.Viper, log *logger.Logger) *messaging.MessagingService {
	rmq, err := messaging.NewRabbitMQ(conf.GetString("messaging.connectionString"), log)
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ", zap.Error(err))
	}

	log.Info("Connected to RabbitMQ")

	return rmq
}

func startHTTPServer(conf *viper.Viper, logger *logger.Logger, rmq *messaging.MessagingService, reconcileJob *reconciliation.Job) {
	validator.InitValidator(logger)

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		},
	})

	// app.Use(swagger.New(
	// 	swagger.Config{
	// 		BasePath: "/",
	// 		FilePath: "./swagger.json",
	// 		Path:     "swagger",
	// 		Title:    "Swagger API Docs",
	// 	},
	// ))

	// CORS middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))

	// JWT Middleware
	jwt := middleware.JWTMiddleware(conf, logger)

	repos := repository.NewFactory()
	services := service.NewServiceFactory(repos, rmq, logger)

	routes.BindHealthRoutes(app, db.GetDB(), rmq)
	routes.BindFlakeRoutes(app, db.GetDB(), jwtpkg.NewJwt(conf), jwt, services)
	routes.BindBalanceRoutes(app, db.GetDB(), jwtpkg.NewJwt(conf), jwt, services)
	routes.BindTransactionRoutes(app, db.GetDB(), jwtpkg.NewJwt(conf), jwt, services)
	routes.BindAdminRoutes(app, jwt, logger, reconcileJob)

	port := conf.GetInt("http.port")
	formattedPort := fmt.Sprintf(":%d", port)

	log.Printf("HTTP server is running on port %d\n", port)
	if err := app.Listen(formattedPort); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}

func startAccountCreationConsumer(rmq *messaging.MessagingService, logger *logger.Logger) {
	exchangeName := "user.created"
	queueName := "user.created.account"

	const (
		initialBackoff = time.Second
		maxBackoff     = 30 * time.Second
	)
	backoff := initialBackoff

	for {
		messages, err := rmq.ConsumeFromExchange(exchangeName, queueName)
		if err != nil {
			logger.Error("Failed to start consumer, retrying",
				zap.Error(err),
				zap.String("queueName", queueName),
				zap.Duration("backoff", backoff),
			)
			time.Sleep(backoff)
			backoff = nextBackoff(backoff, maxBackoff)
			continue
		}

		logger.Info("Consumer started", zap.String("queueName", queueName))
		backoff = initialBackoff

		for msg := range messages {
			logger.Info("Received message",
				zap.String("queueName", queueName),
				zap.String("body", string(msg.Body)),
				zap.String("contentType", msg.ContentType),
			)

			var event UserCreatedEvent
			if err := json.Unmarshal(msg.Body, &event); err != nil {
				logger.Error("Failed to unmarshal message", zap.Error(err), zap.String("body", string(msg.Body)))
				msg.Nack(false, false)
				continue
			}

			if err := processUserCreated(event, logger); err != nil {
				logger.Error("Failed to process user.created event", zap.Error(err), zap.String("userID", event.ID))
				msg.Nack(false, true)
				continue
			}

			msg.Ack(false)
		}

		logger.Warn("Consumer connection lost, reconnecting",
			zap.String("queueName", queueName),
			zap.Duration("backoff", backoff),
		)
		time.Sleep(backoff)
		backoff = nextBackoff(backoff, maxBackoff)
	}
}

type UserCreatedEvent struct {
	ID           string  `json:"id"`
	Username     string  `json:"username"`
	Email        string  `json:"email"`
	AnnualIncome float64 `json:"annual_income"`
	Debt         float64 `json:"debt"`
	AssetsValue  float64 `json:"assets_value"`
	CreatedAt    string  `json:"created_at"`
}

func processUserCreated(event UserCreatedEvent, logger *logger.Logger) error {
	logger.Info("Processing user.created event",
		zap.String("ID", event.ID),
		zap.String("Username", event.Username),
		zap.String("Email", event.Email),
	)

	userID, err := uuid.Parse(event.ID)
	if err != nil {
		return fmt.Errorf("invalid user ID %q: %w", event.ID, err)
	}

	database := db.GetDB()

	// Idempotency: skip if account already exists for this user
	var existing models.Account
	if err := database.Where("user_id = ?", userID).First(&existing).Error; err == nil {
		logger.Info("Account already exists for user, skipping", zap.String("userID", event.ID))
		return nil
	}

	account := models.Account{
		UserID:               userID,
		Balance:              0,
		Status:               models.AccountStatusActive,
		ReconciliationStatus: models.AccountReconciliationStatusOK,
	}

	if err := database.Create(&account).Error; err != nil {
		return fmt.Errorf("failed to create account for user %q: %w", event.ID, err)
	}

	logger.Info("Account created for user",
		zap.String("userID", event.ID),
		zap.String("accountID", account.ID.String()),
	)

	return nil
}

func nextBackoff(current, max time.Duration) time.Duration {
	next := current * 2
	if next > max {
		return max
	}
	return next
}
