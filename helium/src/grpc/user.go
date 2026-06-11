package grpc

import (
	"context"

	"github.com/getsnowflake/snowflake/helium/pb"
	"github.com/getsnowflake/snowflake/helium/src/models"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

type userService struct {
	pb.UnimplementedUserServiceServer
	db *gorm.DB
}

func (s *userService) GetUsers(context.Context, *emptypb.Empty) (*pb.GetUsersResponse, error) {
	var users []models.User

	s.db.Find(&users)

	var grpcUsers []*pb.User
	for _, user := range users {
		grpcUser := user.ToGRPC()
		grpcUsers = append(grpcUsers, grpcUser)
	}

	return &pb.GetUsersResponse{
		Users: grpcUsers,
	}, nil
}

func (s *userService) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	var user models.User

	result := s.db.Where("id = ?", req.Id).First(&user)

	if result.Error != nil {
		return nil, result.Error
	}

	return &pb.GetUserResponse{
		User: user.ToGRPC(),
	}, nil
}

func (s *userService) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	claims := ClaimsFromContext(ctx)
	if claims == nil {
		return nil, status.Error(codes.Unauthenticated, "missing auth claims")
	}

	if claims.Subject != req.Id {
		return nil, status.Error(codes.PermissionDenied, "cannot delete another user's account")
	}

	var user models.User

	result := s.db.Where("id = ?", req.Id).First(&user)

	if result.Error != nil {
		return nil, result.Error
	}

	s.db.Delete(&user)

	return &pb.DeleteUserResponse{
		User: user.ToGRPC(),
	}, nil
}

func RegisterUserService(s *grpc.Server, db *gorm.DB) {
	pb.RegisterUserServiceServer(s, &userService{db: db})
}
