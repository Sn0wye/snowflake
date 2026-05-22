package grpc

import (
	"github.com/getsnowflake/snowflake/helium/src/db"

	grpc "google.golang.org/grpc"
)

func RegisterAllServices(s *grpc.Server) {
	dbInstance := db.GetDB()
	RegisterAuthService(s)
	RegisterUserService(s, dbInstance)
}
