package service

import (
	"github.com/getsnowflake/snowflake/helium/pkg/jwt"
	"github.com/getsnowflake/snowflake/helium/pkg/messaging"
	"github.com/getsnowflake/snowflake/helium/pkg/repository"
)

type Factory struct {
	Auth  AuthService
	OAuth OAuthService
	Token TokenService
}

func NewFactory(repos *repository.Factory, j *jwt.JWT, rmq *messaging.MessagingService) *Factory {
	svc := &Factory{}
	svc.Token = newTokenService(j)
	svc.Auth = newAuthService(repos, j, svc.Token, rmq)
	svc.OAuth = newOAuthService(repos, svc.Token, rmq)
	return svc
}
