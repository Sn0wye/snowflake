package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"snowflake/pkg/config"
	"snowflake/pkg/logger"
	"snowflake/pkg/messaging"
	"snowflake/pkg/validator"
	"snowflake/src/db"
	grpcs "snowflake/src/grpc"
	"snowflake/src/middleware"
	"snowflake/src/migration"
	"snowflake/src/models"
	"snowflake/src/routes"
	"syscall"

	"github.com/gofiber/contrib/swagger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
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

	// Start servers
	go startHTTPServer(conf, logger, rmq)
	go startGRPCServer(conf)

	<-quit // Wait for shutdown signal

	log.Println("Shutting down the servers...")
	log.Println("All servers stopped gracefully")
}

func migrateDB(logger *logger.Logger) {
	db := db.GetDB()

	db.AutoMigrate(
		models.RetrieveAll()...,
	)

	migrate := migration.NewMigrate(
		db,
		logger,
	)

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

func startHTTPServer(conf *viper.Viper, logger *logger.Logger, rmq *messaging.MessagingService) {
	validator.InitValidator(logger)

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		},
	})

	app.Use(swagger.New(
		swagger.Config{
			BasePath: "/",
			FilePath: "./swagger.json",
			Path:     "swagger",
			Title:    "Swagger API Docs",
		},
	))

	// CORS middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))

	// JWT Middleware
	jwt := middleware.JWTMiddleware(conf, logger)

	// Bind routes
	routes.BindAuthRoutes(app, jwt, logger, rmq)

	port := conf.GetInt("http.port")
	formattedPort := fmt.Sprintf(":%d", port)

	log.Printf("HTTP server is running on port %d\n", port)
	if err := app.Listen(formattedPort); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}

func startGRPCServer(conf *viper.Viper) {
	grpcPort := conf.GetInt("grpc.port")
	grpcFormattedPort := fmt.Sprintf(":%d", grpcPort)

	lis, err := net.Listen("tcp", grpcFormattedPort)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	grpcs.RegisterAllServices(s)
	log.Printf("gRPC server is running on port %d\n", grpcPort)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC server: %v", err)
	}
}
