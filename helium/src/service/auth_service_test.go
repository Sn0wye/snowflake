package service

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/getsnowflake/snowflake/helium/pkg/jwt"
	"github.com/getsnowflake/snowflake/helium/pkg/logger"
	"github.com/getsnowflake/snowflake/helium/pkg/messaging"
	"github.com/getsnowflake/snowflake/helium/src/dto"
	"github.com/getsnowflake/snowflake/helium/src/models"
	"github.com/getsnowflake/snowflake/helium/src/repository"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type produceCall struct {
	exchange      string
	message       string
	correlationID string
}

type capturingEventBus struct {
	calls []produceCall
}

func (b *capturingEventBus) ProduceToExchange(exchange, message string) error {
	b.calls = append(b.calls, produceCall{exchange, message, ""})
	return nil
}

func (b *capturingEventBus) ProduceToExchangeWithHeaders(exchange, message, correlationID string) error {
	b.calls = append(b.calls, produceCall{exchange, message, correlationID})
	return nil
}

type authTestFixture struct {
	db       *gorm.DB
	jwter    *jwt.JWT
	auth     AuthService
	tokens   TokenService
	eventBus *capturingEventBus
}

func setupAuthServiceTest(t *testing.T) *authTestFixture {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	if err := db.AutoMigrate(models.RetrieveAll()...); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	jwter, err := jwt.NewTestJWT("test-access-key-32byteslong!!", "test-refresh-key-32bytes!!", "helium-test")
	if err != nil {
		t.Fatalf("failed to create test JWT: %v", err)
	}
	repos := repository.NewFactory()
	bus := &capturingEventBus{}
	log := &logger.Logger{Logger: zap.NewNop()}
	factory := NewFactory(repos, jwter, bus, log)
	return &authTestFixture{
		db:       db,
		jwter:    jwter,
		auth:     factory.Auth,
		tokens:   factory.Token,
		eventBus: bus,
	}
}

func seedUser(t *testing.T, fx *authTestFixture, name, username, email, password string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	user := models.User{
		Name:     name,
		Username: username,
		Email:    email,
		Password: string(hash),
	}
	if err := fx.db.Create(&user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	return user.ID.String()
}

func TestAuthService_Profile_Found(t *testing.T) {
	fx := setupAuthServiceTest(t)
	hash, err := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	user := models.User{
		Name: "Alice", Username: "alice", Email: "alice@test.com",
		Password: string(hash), AnnualIncome: 100000, Debt: 50000, AssetsValue: 200000,
	}
	if err := fx.db.Create(&user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	userID := user.ID.String()

	resp, err := fx.auth.Profile(fx.db, userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != userID {
		t.Fatalf("expected ID %s, got %s", userID, resp.ID)
	}
	if resp.Name != "Alice" {
		t.Fatalf("expected Name Alice, got %s", resp.Name)
	}
	if resp.AnnualIncome != 100000 {
		t.Fatalf("expected AnnualIncome 100000, got %d", resp.AnnualIncome)
	}
	if resp.Debt != 50000 {
		t.Fatalf("expected Debt 50000, got %d", resp.Debt)
	}
	if resp.AssetsValue != 200000 {
		t.Fatalf("expected AssetsValue 200000, got %d", resp.AssetsValue)
	}
}

func TestAuthService_Profile_NotFound(t *testing.T) {
	fx := setupAuthServiceTest(t)
	_, err := fx.auth.Profile(fx.db, "00000000-0000-0000-0000-000000000000")
	if err == nil {
		t.Fatal("expected error for non-existent user")
	}
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestAuthService_Register_Success(t *testing.T) {
	fx := setupAuthServiceTest(t)
	req := dto.RegisterRequest{
		Name:         "Alice",
		Username:     "alice",
		Password:     "secret123",
		Email:        "alice@test.com",
		AnnualIncome: 100000,
		Debt:         50000,
		AssetsValue:  200000,
	}
	resp, err := fx.auth.Register(fx.db, req, zap.NewNop(), "test-correlation-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.AccessToken == "" || resp.RefreshToken == "" {
		t.Fatal("expected non-empty tokens")
	}
	if len(fx.eventBus.calls) != 1 {
		t.Fatalf("expected 1 event, got %d", len(fx.eventBus.calls))
	}
	call := fx.eventBus.calls[0]
	if call.exchange != messaging.ExchangeUserCreated {
		t.Fatalf("expected exchange %s, got %s", messaging.ExchangeUserCreated, call.exchange)
	}
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(call.message), &payload); err != nil {
		t.Fatalf("failed to unmarshal event payload: %v", err)
	}
	if payload["username"] != "alice" {
		t.Fatalf("expected username alice, got %v", payload["username"])
	}
	if payload["annual_income"] != float64(100000) {
		t.Fatalf("expected annual_income 100000, got %v", payload["annual_income"])
	}
	if call.correlationID != "test-correlation-id" {
		t.Fatalf("expected correlationID test-correlation-id, got %s", call.correlationID)
	}
}

func TestAuthService_Register_EmailAlreadyTaken(t *testing.T) {
	fx := setupAuthServiceTest(t)
	req := dto.RegisterRequest{
		Name: "Alice", Username: "alice", Password: "secret123",
		Email: "alice@test.com",
	}
	_, err := fx.auth.Register(fx.db, req, zap.NewNop(), "")
	if err != nil {
		t.Fatalf("first register should succeed: %v", err)
	}
	_, err = fx.auth.Register(fx.db, req, zap.NewNop(), "")
	if err == nil {
		t.Fatal("expected error for duplicate email")
	}
	if !errors.Is(err, ErrEmailAlreadyTaken) {
		t.Fatalf("expected ErrEmailAlreadyTaken, got %v", err)
	}
}

func TestAuthService_Register_NilRMQNoPanic(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	if err := db.AutoMigrate(models.RetrieveAll()...); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	jwter, _ := jwt.NewTestJWT("test-access-key-32byteslong!!", "test-refresh-key-32bytes!!", "helium-test")
	repos := repository.NewFactory()
	log := &logger.Logger{Logger: zap.NewNop()}
	factory := NewFactory(repos, jwter, nil, log)
	req := dto.RegisterRequest{
		Name: "Alice", Username: "alice", Password: "secret123",
		Email: "alice@test.com",
	}
	resp, err := factory.Auth.Register(db, req, zap.NewNop(), "")
	if err != nil {
		t.Fatalf("register should succeed with nil rmq: %v", err)
	}
	if resp.AccessToken == "" {
		t.Fatal("expected access token")
	}
}

func TestAuthService_Login_Success(t *testing.T) {
	fx := setupAuthServiceTest(t)
	seedUser(t, fx, "Bob", "bob", "bob@test.com", "secret123")

	resp, err := fx.auth.Login(fx.db, dto.LoginRequest{
		Email: "bob@test.com", Password: "secret123",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.AccessToken == "" || resp.RefreshToken == "" {
		t.Fatal("expected non-empty tokens")
	}
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	fx := setupAuthServiceTest(t)
	_, err := fx.auth.Login(fx.db, dto.LoginRequest{
		Email: "ghost@test.com", Password: "whatever",
	})
	if err == nil {
		t.Fatal("expected error for unknown user")
	}
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	fx := setupAuthServiceTest(t)
	seedUser(t, fx, "Bob", "bob", "bob@test.com", "correct")

	_, err := fx.auth.Login(fx.db, dto.LoginRequest{
		Email: "bob@test.com", Password: "wrong",
	})
	if err == nil {
		t.Fatal("expected error for wrong password")
	}
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthService_Refresh_Success(t *testing.T) {
	fx := setupAuthServiceTest(t)
	userID := seedUser(t, fx, "Eve", "eve", "eve@test.com", "secret")
	refreshToken, err := fx.tokens.GenerateRefreshToken(fx.db, userID)
	if err != nil {
		t.Fatalf("failed to generate refresh token: %v", err)
	}
	resp, err := fx.auth.Refresh(fx.db, refreshToken)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.AccessToken == "" || resp.RefreshToken == "" {
		t.Fatal("expected non-empty new tokens")
	}
	_, err = fx.jwter.ParseRefreshToken(resp.RefreshToken)
	if err != nil {
		t.Fatalf("new refresh token should be valid, got: %v", err)
	}
}

func TestAuthService_Refresh_InvalidToken(t *testing.T) {
	fx := setupAuthServiceTest(t)
	_, err := fx.auth.Refresh(fx.db, "garbage-token")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthService_Refresh_TokenNotFound(t *testing.T) {
	fx := setupAuthServiceTest(t)
	userID := seedUser(t, fx, "Eve", "eve", "eve@test.com", "secret")
	token, err := fx.jwter.GenRefreshToken(userID, time.Now().Add(30*24*time.Hour))
	if err != nil {
		t.Fatalf("failed to generate jwt refresh token: %v", err)
	}
	_, err = fx.auth.Refresh(fx.db, token)
	if err == nil {
		t.Fatal("expected error for token not found in DB")
	}
	if !errors.Is(err, ErrRefreshTokenNotFound) {
		t.Fatalf("expected ErrRefreshTokenNotFound, got %v", err)
	}
}

func TestAuthService_Logout_Success(t *testing.T) {
	fx := setupAuthServiceTest(t)
	userID := seedUser(t, fx, "Dave", "dave", "dave@test.com", "secret")
	refreshToken, err := fx.tokens.GenerateRefreshToken(fx.db, userID)
	if err != nil {
		t.Fatalf("failed to generate refresh token: %v", err)
	}
	if err := fx.auth.Logout(fx.db, refreshToken); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAuthService_Logout_DoubleLogout(t *testing.T) {
	fx := setupAuthServiceTest(t)
	userID := seedUser(t, fx, "Dave", "dave", "dave@test.com", "secret")
	refreshToken, err := fx.tokens.GenerateRefreshToken(fx.db, userID)
	if err != nil {
		t.Fatalf("failed to generate refresh token: %v", err)
	}
	if err := fx.auth.Logout(fx.db, refreshToken); err != nil {
		t.Fatalf("first logout should succeed: %v", err)
	}
	err = fx.auth.Logout(fx.db, refreshToken)
	if err == nil {
		t.Fatal("expected error on second logout")
	}
	if !errors.Is(err, ErrRefreshTokenNotFound) {
		t.Fatalf("expected ErrRefreshTokenNotFound, got %v", err)
	}
}

func TestAuthService_Logout_InvalidToken(t *testing.T) {
	fx := setupAuthServiceTest(t)
	err := fx.auth.Logout(fx.db, "garbage")
	if err == nil {
		t.Fatal("expected error for non-existent token hash")
	}
	if !errors.Is(err, ErrRefreshTokenNotFound) {
		t.Fatalf("expected ErrRefreshTokenNotFound, got %v", err)
	}
}
