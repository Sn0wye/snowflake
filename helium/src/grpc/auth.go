package grpc

import (
	"context"
	"snowflake/pb"
	"snowflake/pkg/config"
	"snowflake/pkg/jwt"
	"time"

	"google.golang.org/grpc"
)

type authService struct {
	pb.UnimplementedAuthServiceServer
}

func (s *authService) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	conf := config.GetConfig()
	jwt := jwt.NewJwt(conf)

	_, err := jwt.ParseToken(req.Token)

	if err != nil {
		return &pb.ValidateTokenResponse{Valid: false}, err
	}

	return &pb.ValidateTokenResponse{Valid: true}, nil
}

func (s *authService) ParseToken(ctx context.Context, req *pb.ParseTokenRequest) (*pb.ParseTokenResponse, error) {
	conf := config.GetConfig()
	jwt := jwt.NewJwt(conf)

	claims, err := jwt.ParseToken(req.Token)

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
	pb.RegisterAuthServiceServer(s, &authService{})
}
