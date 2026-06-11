package grpc

import (
	"github.com/getsnowflake/snowflake/helium/pkg/jwt"
	"github.com/getsnowflake/snowflake/helium/src/db"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func NewServer(log *zap.Logger) *grpc.Server {
	jwter := newJWT()
	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			RecoveryInterceptor(log),
			LoggingInterceptor(log),
			AuthInterceptor(jwter),
		),
	)
	RegisterAllServices(s, jwter)
	return s
}

func RegisterAllServices(s *grpc.Server, jwter *jwt.JWT) {
	dbInstance := db.GetDB()
	RegisterAuthService(s, jwter)
	RegisterUserService(s, dbInstance)
}
