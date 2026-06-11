package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/getsnowflake/snowflake/helium/pb"
	"github.com/getsnowflake/snowflake/helium/pkg/jwt"
)

func newTestAuthService(t *testing.T) *authService {
	t.Helper()
	jwter, err := jwt.NewTestJWT("test-access-key-32byteslong!!", "test-refresh-key-32bytes!!", "helium-test")
	if err != nil {
		t.Fatalf("failed to create test JWT: %v", err)
	}
	return &authService{jwter: jwter}
}

func TestValidateToken_Valid(t *testing.T) {
	svc := newTestAuthService(t)
	token, _ := svc.jwter.GenToken("user-1", time.Now().Add(time.Hour))
	resp, err := svc.ValidateToken(context.Background(), &pb.ValidateTokenRequest{Token: token})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Valid {
		t.Fatal("expected Valid=true for valid token")
	}
}

func TestValidateToken_Expired(t *testing.T) {
	svc := newTestAuthService(t)
	token, _ := svc.jwter.GenToken("user-1", time.Now().Add(-time.Hour))
	resp, err := svc.ValidateToken(context.Background(), &pb.ValidateTokenRequest{Token: token})
	if err == nil {
		t.Fatal("expected error for expired token alongside response")
	}
	if resp.Valid {
		t.Fatal("expected Valid=false for expired token")
	}
}

func TestValidateToken_Garbage(t *testing.T) {
	svc := newTestAuthService(t)
	resp, err := svc.ValidateToken(context.Background(), &pb.ValidateTokenRequest{Token: "garbage"})
	if err == nil {
		t.Fatal("expected error for garbage token alongside response")
	}
	if resp.Valid {
		t.Fatal("expected Valid=false for garbage token")
	}
}

func TestParseToken_Valid(t *testing.T) {
	svc := newTestAuthService(t)
	token, _ := svc.jwter.GenToken("user-1", time.Now().Add(time.Hour))
	resp, err := svc.ParseToken(context.Background(), &pb.ParseTokenRequest{Token: token})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Sub != "user-1" {
		t.Fatalf("expected subject user-1, got %s", resp.Sub)
	}
	if resp.Iss != "helium-test" {
		t.Fatalf("expected issuer helium-test, got %s", resp.Iss)
	}
	if resp.Iat == "" {
		t.Fatal("expected non-empty issued-at")
	}
	if resp.Exp == "" {
		t.Fatal("expected non-empty expires-at")
	}
}

func TestParseToken_Invalid(t *testing.T) {
	svc := newTestAuthService(t)
	_, err := svc.ParseToken(context.Background(), &pb.ParseTokenRequest{Token: "not-a-token"})
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}

func TestParseToken_ReturnsRFC3339Timestamps(t *testing.T) {
	svc := newTestAuthService(t)
	token, _ := svc.jwter.GenToken("user-1", time.Now().Add(time.Hour))
	resp, _ := svc.ParseToken(context.Background(), &pb.ParseTokenRequest{Token: token})
	_, err := time.Parse(time.RFC3339, resp.Iat)
	if err != nil {
		t.Fatalf("Iat is not valid RFC3339: %s", resp.Iat)
	}
	_, err = time.Parse(time.RFC3339, resp.Exp)
	if err != nil {
		t.Fatalf("Exp is not valid RFC3339: %s", resp.Exp)
	}
}

func TestParseToken_BearerPrefix(t *testing.T) {
	svc := newTestAuthService(t)
	token, _ := svc.jwter.GenToken("user-1", time.Now().Add(time.Hour))
	resp, err := svc.ParseToken(context.Background(), &pb.ParseTokenRequest{Token: "Bearer " + token})
	if err != nil {
		t.Fatalf("unexpected error with Bearer prefix: %v", err)
	}
	if resp.Sub != "user-1" {
		t.Fatalf("expected subject user-1, got %s", resp.Sub)
	}
}

func TestValidateToken_ReturnsValidFalseOnError(t *testing.T) {
	svc := newTestAuthService(t)
	resp, err := svc.ValidateToken(context.Background(), &pb.ValidateTokenRequest{Token: "junk"})
	if err == nil {
		t.Fatal("expected error alongside response for invalid token")
	}
	if resp.Valid {
		t.Fatal("expected Valid=false for invalid token")
	}
}
