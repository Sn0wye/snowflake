package service

import (
	"github.com/getsnowflake/snowflake/helium/pkg/jwt"
	"github.com/getsnowflake/snowflake/helium/pkg/logger"
	"github.com/getsnowflake/snowflake/helium/pkg/messaging"
	"github.com/getsnowflake/snowflake/helium/pkg/repository"
)

type Factory struct {
	Auth  AuthService
	OAuth OAuthService
	Token TokenService
}

func NewFactory(repos *repository.Factory, j *jwt.JWT, rmq *messaging.MessagingService, log *logger.Logger) *Factory {
	svc := &Factory{}
	svc.Token = newTokenService(j, repos.RefreshToken)
	svc.Auth = newAuthService(repos, j, svc.Token, rmq, log)
	svc.OAuth = newOAuthService(repos, svc.Token, rmq, log)
	return svc
}
