package service

import (
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/getsnowflake/snowflake/helium/pkg/jwt"
	"github.com/getsnowflake/snowflake/helium/src/models"
	"github.com/getsnowflake/snowflake/helium/src/repository"
)

func setupTokenTest(t *testing.T) (*gorm.DB, *jwt.JWT, TokenService) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.RefreshToken{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	jwter := setupTestJWT(t)
	repo := repository.NewRefreshTokenRepo()
	svc := newTokenService(jwter, repo)
	return db, jwter, svc
}

func setupTestJWT(t *testing.T) *jwt.JWT {
	t.Helper()
	j, err := jwt.NewTestJWT("test-access-key-32byteslong!!", "test-refresh-key-32bytes!!", "helium-test")
	if err != nil {
		t.Fatalf("failed to create test JWT: %v", err)
	}
	return j
}

func TestGenerateAccessToken_ReturnsToken(t *testing.T) {
	_, _, svc := setupTokenTest(t)
	token, err := svc.GenerateAccessToken("1b7e4a5e-9d3c-4e2f-8a1d-6c5b9e0f1a2b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty access token")
	}
}

func TestGenerateAccessToken_InvalidUserID_StillWorks(t *testing.T) {
	_, _, svc := setupTokenTest(t)
	token, err := svc.GenerateAccessToken("random-string")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty access token")
	}
}

func TestGenerateRefreshToken_Success(t *testing.T) {
	db, jwter, svc := setupTokenTest(t)
	token, err := svc.GenerateRefreshToken(db, "1b7e4a5e-9d3c-4e2f-8a1d-6c5b9e0f1a2b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty refresh token")
	}
	claims, err := jwter.ParseRefreshToken(token)
	if err != nil {
		t.Fatalf("failed to parse generated refresh token: %v", err)
	}
	if claims.Subject != "1b7e4a5e-9d3c-4e2f-8a1d-6c5b9e0f1a2b" {
		t.Fatalf("expected subject, got %s", claims.Subject)
	}
}

func TestGenerateRefreshToken_InvalidUserID(t *testing.T) {
	db, _, svc := setupTokenTest(t)
	_, err := svc.GenerateRefreshToken(db, "not-a-uuid")
	if err == nil {
		t.Fatal("expected error for invalid user ID")
	}
}

func TestGenerateRefreshToken_NotExpired(t *testing.T) {
	db, jwter, svc := setupTokenTest(t)
	token, err := svc.GenerateRefreshToken(db, "1b7e4a5e-9d3c-4e2f-8a1d-6c5b9e0f1a2b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	claims, err := jwter.ParseRefreshToken(token)
	if err != nil {
		t.Fatalf("expected refresh token to be valid and not expired, got: %v", err)
	}
	now := time.Now()
	if claims.ExpiresAt == nil || !claims.ExpiresAt.After(now) {
		t.Fatal("refresh token should expire in the future")
	}
	if claims.ExpiresAt.After(now.Add(refreshTokenDuration + time.Minute)) {
		t.Fatal("refresh token expiry exceeds expected duration")
	}
}

func TestRevokeAllUserRefreshTokens_Success(t *testing.T) {
	db, _, svc := setupTokenTest(t)
	_, err := svc.GenerateRefreshToken(db, "1b7e4a5e-9d3c-4e2f-8a1d-6c5b9e0f1a2b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = svc.GenerateRefreshToken(db, "a1b2c3d4-e5f6-7890-abcd-ef1234567890")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = svc.RevokeAllUserRefreshTokens(db, "1b7e4a5e-9d3c-4e2f-8a1d-6c5b9e0f1a2b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = svc.GenerateRefreshToken(db, "1b7e4a5e-9d3c-4e2f-8a1d-6c5b9e0f1a2b")
	if err != nil {
		t.Fatalf("should still be able to generate new tokens after revoke: %v", err)
	}
}

func TestRevokeAllUserRefreshTokens_InvalidUserID(t *testing.T) {
	db, _, svc := setupTokenTest(t)
	err := svc.RevokeAllUserRefreshTokens(db, "not-a-uuid")
	if err == nil {
		t.Fatal("expected error for invalid user ID")
	}
}

func TestRevokeAllUserRefreshTokens_NoTokens(t *testing.T) {
	db, _, svc := setupTokenTest(t)
	err := svc.RevokeAllUserRefreshTokens(db, "1b7e4a5e-9d3c-4e2f-8a1d-6c5b9e0f1a2b")
	if err != nil {
		t.Fatalf("unexpected error when no tokens exist: %v", err)
	}
}

func TestHashToken_Deterministic(t *testing.T) {
	h1 := hashToken("test-token")
	h2 := hashToken("test-token")
	if h1 != h2 {
		t.Fatal("hash should be deterministic")
	}
	if len(h1) != 64 {
		t.Fatalf("SHA-256 hex should be 64 chars, got %d", len(h1))
	}
}

func TestHashToken_DifferentInputs(t *testing.T) {
	h1 := hashToken("token-a")
	h2 := hashToken("token-b")
	if h1 == h2 {
		t.Fatal("different tokens should produce different hashes")
	}
}

func TestAccessTokenDuration_Is15Minutes(t *testing.T) {
	if accessTokenDuration != 15*time.Minute {
		t.Fatalf("expected 15 minutes, got %v", accessTokenDuration)
	}
}

func TestRefreshTokenDuration_Is30Days(t *testing.T) {
	expected := 30 * 24 * time.Hour
	if refreshTokenDuration != expected {
		t.Fatalf("expected %v, got %v", expected, refreshTokenDuration)
	}
}
