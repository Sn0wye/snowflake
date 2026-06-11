package controllers

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/getsnowflake/snowflake/helium/pkg/jwt"
	"github.com/getsnowflake/snowflake/helium/pkg/validator"
	"github.com/getsnowflake/snowflake/helium/src/dto"
	"github.com/getsnowflake/snowflake/helium/src/models"
	"github.com/getsnowflake/snowflake/helium/src/repository"
	"github.com/getsnowflake/snowflake/helium/src/service"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type controllerFixture struct {
	app    *fiber.App
	ctrl   AuthController
	db     *gorm.DB
	jwter  *jwt.JWT
	tokens service.TokenService
}

func setupControllerTest(t *testing.T) *controllerFixture {
	t.Helper()
	validator.InitValidatorForTest()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	if err := db.AutoMigrate(models.RetrieveAll()...); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	jwter, _ := jwt.NewTestJWT("test-access-key-32byteslong!!", "test-refresh-key-32bytes!!", "helium-test")
	repos := repository.NewFactory()
	svcFactory := service.NewFactory(repos, jwter, nil, nil)
	ctrl := NewAuthController(db, jwter, svcFactory.Auth, nil)
	app := fiber.New()
	return &controllerFixture{
		app:    app,
		ctrl:   ctrl,
		db:     db,
		jwter:  jwter,
		tokens: svcFactory.Token,
	}
}

func (fx *controllerFixture) seedUser(email, password string) *models.User {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	user := models.User{
		Name: email, Username: email, Email: email, Password: string(hash),
	}
	if err := fx.db.Create(&user).Error; err != nil {
		panic(err)
	}
	return &user
}

func (fx *controllerFixture) seedRefreshToken(userID string) string {
	token, err := fx.tokens.GenerateRefreshToken(fx.db, userID)
	if err != nil {
		panic(err)
	}
	return token
}

func (fx *controllerFixture) registerRoute(body string) *http.Response {
	fx.app.Post("/auth/register", func(c *fiber.Ctx) error { return fx.ctrl.Register(c) })
	req := httptest.NewRequest("POST", "/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := fx.app.Test(req)
	return resp
}

func (fx *controllerFixture) loginRoute(body string) *http.Response {
	fx.app.Post("/auth/login", func(c *fiber.Ctx) error { return fx.ctrl.Login(c) })
	req := httptest.NewRequest("POST", "/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := fx.app.Test(req)
	return resp
}

func (fx *controllerFixture) refreshRoute(body string) *http.Response {
	fx.app.Post("/auth/refresh", func(c *fiber.Ctx) error { return fx.ctrl.Refresh(c) })
	req := httptest.NewRequest("POST", "/auth/refresh", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := fx.app.Test(req)
	return resp
}

func (fx *controllerFixture) logoutRoute(body string) *http.Response {
	fx.app.Post("/auth/logout", func(c *fiber.Ctx) error { return fx.ctrl.Logout(c) })
	req := httptest.NewRequest("POST", "/auth/logout", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := fx.app.Test(req)
	return resp
}

func makeRegisterBody(name, username, password, email string) string {
	return `{"name":"` + name + `","username":"` + username + `","password":"` + password +
		`","email":"` + email + `","annual_income":0,"debt":0,"assets_value":0}`
}

func TestControllerRegister_Success(t *testing.T) {
	fx := setupControllerTest(t)
	body := `{"name":"Alice","username":"alice","password":"password123","email":"alice@test.com","annual_income":100000,"debt":50000,"assets_value":200000}`
	resp := fx.registerRoute(body)
	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d. body: %s", resp.StatusCode, string(bodyBytes))
	}
	var regResp dto.RegisterResponse
	if err := json.Unmarshal(bodyBytes, &regResp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if regResp.AccessToken == "" || regResp.RefreshToken == "" {
		t.Fatal("expected access_token and refresh_token")
	}
}

func TestControllerRegister_DuplicateEmail(t *testing.T) {
	fx := setupControllerTest(t)
	body := makeRegisterBody("Alice", "alice", "pass123", "dup@test.com")
	resp := fx.registerRoute(body)
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("first register failed with %d", resp.StatusCode)
	}
	resp2 := fx.registerRoute(body)
	if resp2.StatusCode != fiber.StatusConflict {
		t.Fatalf("expected 409 for duplicate email, got %d", resp2.StatusCode)
	}
}

func TestControllerLogin_Success(t *testing.T) {
	fx := setupControllerTest(t)
	fx.seedUser("bob@test.com", "secret123")
	resp := fx.loginRoute(`{"email":"bob@test.com","password":"secret123"}`)
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d. body: %s", resp.StatusCode, string(body))
	}
	var loginResp dto.LoginResponse
	json.Unmarshal(body, &loginResp)
	if loginResp.AccessToken == "" {
		t.Fatal("expected access_token in login response")
	}
}

func TestControllerLogin_WrongPassword(t *testing.T) {
	fx := setupControllerTest(t)
	fx.seedUser("bob2@test.com", "correct")
	resp := fx.loginRoute(`{"email":"bob2@test.com","password":"wrongpassword"}`)
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected 401 for wrong password, got %d", resp.StatusCode)
	}
}

func TestControllerLogin_NonexistentUser(t *testing.T) {
	fx := setupControllerTest(t)
	resp := fx.loginRoute(`{"email":"ghost@test.com","password":"whatever"}`)
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected 401 for non-existent user, got %d", resp.StatusCode)
	}
}

func TestControllerRefresh_Success(t *testing.T) {
	fx := setupControllerTest(t)
	user := fx.seedUser("eve@test.com", "pass")
	token := fx.seedRefreshToken(user.ID.String())
	resp := fx.refreshRoute(`{"refresh_token":"` + token + `"}`)
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 for refresh, got %d. body: %s", resp.StatusCode, string(body))
	}
	var refreshResp dto.RefreshResponse
	json.Unmarshal(body, &refreshResp)
	if refreshResp.AccessToken == "" || refreshResp.RefreshToken == "" {
		t.Fatal("expected access_token and refresh_token in refresh response")
	}
}

func TestControllerRefresh_InvalidToken(t *testing.T) {
	fx := setupControllerTest(t)
	resp := fx.refreshRoute(`{"refresh_token":"garbage"}`)
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected 401 for invalid refresh token, got %d", resp.StatusCode)
	}
}

func TestControllerLogout_Success(t *testing.T) {
	fx := setupControllerTest(t)
	user := fx.seedUser("dave@test.com", "pass")
	token := fx.seedRefreshToken(user.ID.String())
	resp := fx.logoutRoute(`{"refresh_token":"` + token + `"}`)
	if resp.StatusCode != fiber.StatusNoContent {
		t.Fatalf("expected 204 for logout, got %d", resp.StatusCode)
	}
}

func TestControllerLogout_DoubleLogout(t *testing.T) {
	fx := setupControllerTest(t)
	user := fx.seedUser("frank@test.com", "pass")
	token := fx.seedRefreshToken(user.ID.String())
	resp1 := fx.logoutRoute(`{"refresh_token":"` + token + `"}`)
	if resp1.StatusCode != fiber.StatusNoContent {
		t.Fatalf("expected 204 for first logout, got %d", resp1.StatusCode)
	}
	resp2 := fx.logoutRoute(`{"refresh_token":"` + token + `"}`)
	if resp2.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected 401 for double logout, got %d", resp2.StatusCode)
	}
}

func TestControllerLogout_InvalidToken(t *testing.T) {
	fx := setupControllerTest(t)
	resp := fx.logoutRoute(`{"refresh_token":"invalid"}`)
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected 401 for invalid token on logout, got %d", resp.StatusCode)
	}
}
