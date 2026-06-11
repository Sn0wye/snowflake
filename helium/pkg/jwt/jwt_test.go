package jwt

import (
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"
)

func newTestJWT(t *testing.T) *JWT {
	t.Helper()
	v := viper.New()
	v.Set("security.jwt_secret", "test-access-key-32byteslong!!")
	v.Set("security.refresh_secret", "test-refresh-key-32bytes!!")
	v.Set("security.jwt_issuer", "helium-test")
	return NewJwt(v)
}

func TestGenToken_ReturnsSignedString(t *testing.T) {
	j := newTestJWT(t)
	token, err := j.GenToken("user-1", time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token string")
	}
	if !strings.Contains(token, ".") {
		t.Fatalf("expected JWT format with dots, got: %s", token)
	}
}

func TestGenToken_DifferentUserIDs(t *testing.T) {
	j := newTestJWT(t)
	t1, _ := j.GenToken("user-a", time.Now().Add(time.Hour))
	t2, _ := j.GenToken("user-b", time.Now().Add(time.Hour))
	if t1 == t2 {
		t.Fatal("tokens for different users should differ")
	}
}

func TestGenRefreshToken_UsesRefreshKey(t *testing.T) {
	j := newTestJWT(t)
	token, err := j.GenRefreshToken("user-1", time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = j.ParseToken(token)
	if err == nil {
		t.Fatal("refresh token should not be parseable with access key")
	}
	claims, err := j.ParseRefreshToken(token)
	if err != nil {
		t.Fatalf("expected refresh token to be parseable: %v", err)
	}
	if claims.Subject != "user-1" {
		t.Fatalf("expected subject user-1, got %s", claims.Subject)
	}
}

func TestParseToken_Valid(t *testing.T) {
	j := newTestJWT(t)
	exp := time.Now().Add(time.Hour)
	token, _ := j.GenToken("user-1", exp)
	claims, err := j.ParseToken(token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if claims.Subject != "user-1" {
		t.Fatalf("expected subject user-1, got %s", claims.Subject)
	}
	if claims.Issuer != "helium-test" {
		t.Fatalf("expected issuer helium-test, got %s", claims.Issuer)
	}
}

func TestParseToken_StripsBearerPrefix(t *testing.T) {
	j := newTestJWT(t)
	token, _ := j.GenToken("user-1", time.Now().Add(time.Hour))
	claims, err := j.ParseToken("Bearer " + token)
	if err != nil {
		t.Fatalf("unexpected error with Bearer prefix: %v", err)
	}
	if claims.Subject != "user-1" {
		t.Fatalf("expected subject user-1, got %s", claims.Subject)
	}
}

func TestParseToken_Expired(t *testing.T) {
	j := newTestJWT(t)
	token, _ := j.GenToken("user-1", time.Now().Add(-time.Hour))
	_, err := j.ParseToken(token)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
	if !strings.Contains(err.Error(), "expired") {
		t.Fatalf("expected expired error, got: %v", err)
	}
}

func TestParseToken_WrongIssuer(t *testing.T) {
	j := newTestJWT(t)
	token, _ := j.GenToken("user-1", time.Now().Add(time.Hour))
	j2 := newTestJWT(t)
	j2.issuer = "other-issuer"
	_, err := j2.ParseToken(token)
	if err == nil {
		t.Fatal("expected error for wrong issuer")
	}
	if !strings.Contains(err.Error(), "issuer") {
		t.Fatalf("expected issuer error, got: %v", err)
	}
}

func TestParseToken_WrongKey(t *testing.T) {
	j := newTestJWT(t)
	token, _ := j.GenToken("user-1", time.Now().Add(time.Hour))
	j2 := newTestJWT(t)
	j2.key = []byte("completely-different-access-key!!")
	_, err := j2.ParseToken(token)
	if err == nil {
		t.Fatal("expected error for token signed with different key")
	}
}

func TestParseToken_GarbageInput(t *testing.T) {
	j := newTestJWT(t)
	_, err := j.ParseToken("not-a-valid-token")
	if err == nil {
		t.Fatal("expected error for garbage input")
	}
}

func TestParseRefreshToken_Valid(t *testing.T) {
	j := newTestJWT(t)
	token, _ := j.GenRefreshToken("user-1", time.Now().Add(time.Hour))
	claims, err := j.ParseRefreshToken(token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if claims.Subject != "user-1" {
		t.Fatalf("expected subject user-1, got %s", claims.Subject)
	}
}

func TestParseRefreshToken_Expired(t *testing.T) {
	j := newTestJWT(t)
	token, _ := j.GenRefreshToken("user-1", time.Now().Add(-time.Hour))
	_, err := j.ParseRefreshToken(token)
	if err == nil {
		t.Fatal("expected error for expired refresh token")
	}
	if !strings.Contains(err.Error(), "expired") {
		t.Fatalf("expected expired error, got: %v", err)
	}
}

func TestNewJwt_PanicsOnDefaultAccessSecret(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for default access secret")
		}
	}()
	v := viper.New()
	v.Set("security.jwt_secret", "change-me-in-production-jwt-secret")
	v.Set("security.refresh_secret", "some-other-secret-thats-long-enough")
	v.Set("security.jwt_issuer", "test")
	_ = NewJwt(v)
}

func TestNewJwt_PanicsOnDefaultRefreshSecret(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for default refresh secret")
		}
	}()
	v := viper.New()
	v.Set("security.jwt_secret", "some-other-secret-thats-long-enough")
	v.Set("security.refresh_secret", "change-me-in-production-refresh-secret")
	v.Set("security.jwt_issuer", "test")
	_ = NewJwt(v)
}

func TestKey_ReturnsAccessKey(t *testing.T) {
	j := newTestJWT(t)
	key := j.Key()
	if string(key) != "test-access-key-32byteslong!!" {
		t.Fatalf("expected access key, got %s", string(key))
	}
}

func TestIssuer_ReturnsIssuer(t *testing.T) {
	j := newTestJWT(t)
	if j.Issuer() != "helium-test" {
		t.Fatalf("expected helium-test issuer, got %s", j.Issuer())
	}
}
