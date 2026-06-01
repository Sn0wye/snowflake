package jwt

import (
	"errors"
	"strings"
	"time"

	jwtLib "github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
)

type JWT struct {
	key        []byte
	refreshKey []byte
	issuer     string
}

type Claims struct {
	jwtLib.RegisteredClaims
}

func NewJwt(conf *viper.Viper) *JWT {
	if conf.GetString("security.jwt_secret") == "change-me-in-production-jwt-secret" {
		panic("security.jwt_secret is set to insecure default 'change-me-in-production-jwt-secret'. Set JWT_SECRET env var.")
	}
	if conf.GetString("security.refresh_secret") == "change-me-in-production-refresh-secret" {
		panic("security.refresh_secret is set to insecure default 'change-me-in-production-refresh-secret'. Set REFRESH_TOKEN_SECRET env var.")
	}

	return &JWT{
		key:        []byte(conf.GetString("security.jwt_secret")),
		refreshKey: []byte(conf.GetString("security.refresh_secret")),
		issuer:     conf.GetString("security.jwt_issuer"),
	}
}

func (j *JWT) Key() []byte {
	return j.key
}

func (j *JWT) Issuer() string {
	return j.issuer
}

func (j *JWT) GenToken(userId string, expiresAt time.Time) (string, error) {
	claims := Claims{
		RegisteredClaims: jwtLib.RegisteredClaims{
			ExpiresAt: jwtLib.NewNumericDate(expiresAt),
			IssuedAt:  jwtLib.NewNumericDate(time.Now()),
			Issuer:    j.issuer,
			Subject:   userId,
		},
	}

	token := jwtLib.NewWithClaims(jwtLib.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(j.key)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (j *JWT) GenRefreshToken(userId string, expiresAt time.Time) (string, error) {
	claims := Claims{
		RegisteredClaims: jwtLib.RegisteredClaims{
			ExpiresAt: jwtLib.NewNumericDate(expiresAt),
			IssuedAt:  jwtLib.NewNumericDate(time.Now()),
			Issuer:    j.issuer,
			Subject:   userId,
		},
	}

	token := jwtLib.NewWithClaims(jwtLib.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(j.refreshKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (j *JWT) ParseRefreshToken(tokenString string) (*Claims, error) {
	token, err := jwtLib.ParseWithClaims(tokenString, &Claims{}, func(token *jwtLib.Token) (interface{}, error) {
		return j.refreshKey, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("token claims are invalid")
	}

	if time.Now().After(claims.ExpiresAt.Time) {
		return nil, errors.New("token is expired")
	}

	if claims.Issuer != j.issuer {
		return nil, errors.New("token issuer is invalid")
	}

	return claims, nil
}

func (j *JWT) ParseToken(tokenString string) (*Claims, error) {
	strippedToken := strings.TrimPrefix(tokenString, "Bearer ")
	token, err := jwtLib.ParseWithClaims(strippedToken, &Claims{}, func(token *jwtLib.Token) (interface{}, error) {
		return j.key, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("token claims are invalid")
	}

	if time.Now().After(claims.ExpiresAt.Time) {
		return nil, errors.New("token is expired")
	}

	if claims.Issuer != j.issuer {
		return nil, errors.New("token issuer is invalid")
	}

	return claims, nil
}
