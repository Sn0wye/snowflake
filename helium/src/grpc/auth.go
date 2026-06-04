package grpc

import (
	"context"
	"time"

	"github.com/getsnowflake/snowflake/helium/pb"
	"github.com/getsnowflake/snowflake/helium/pkg/config"
	"github.com/getsnowflake/snowflake/helium/pkg/jwt"

	"google.golang.org/grpc"
)

type authService struct {
	pb.UnimplementedAuthServiceServer
	jwter *jwt.JWT
}

func (s *authService) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	_, err := s.jwter.ParseToken(req.Token)

	if err != nil {
		return &pb.ValidateTokenResponse{Valid: false}, err
	}

	return &pb.ValidateTokenResponse{Valid: true}, nil
}

func (s *authService) ParseToken(ctx context.Context, req *pb.ParseTokenRequest) (*pb.ParseTokenResponse, error) {
	claims, err := s.jwter.ParseToken(req.Token)

	if err != nil {
		return nil, err
	}

	return &pb.ParseTokenResponse{
		Sub: claims.Subject,
		Iss: claims.Issuer,
		Iat: claims.IssuedAt.Format(time.RFC3339),
		Exp: claims.ExpiresAt.Format(time.RFC3339),
	}, nil
}

func RegisterAuthService(s *grpc.Server) {
	jwter := jwt.NewJwt(config.GetConfig())
	pb.RegisterAuthServiceServer(s, &authService{jwter: jwter})
}
