package grpc

import (
	"context"
	"snowflake/pb"
	"snowflake/src/db"
	"snowflake/src/models"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type userSevice struct {
	pb.UnimplementedUserServiceServer
}

func (s *userSevice) GetUsers(context.Context, *emptypb.Empty) (*pb.GetUsersResponse, error) {
	db := db.GetDB()
	var users []models.User

	db.Find(&users)

	var grpcUsers []*pb.User
	for _, user := range users {
		grpcUser := user.ToGRPC()
		grpcUsers = append(grpcUsers, grpcUser)
	}

	return &pb.GetUsersResponse{
		Users: grpcUsers,
	}, nil
}

func (s *userSevice) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	db := db.GetDB()
	var user models.User

	result := db.Where("id = ?", req.Id).First(&user)

	if result.Error != nil {
		return nil, result.Error
	}

	return &pb.GetUserResponse{
		User: user.ToGRPC(),
	}, nil
}

func (s *userSevice) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	db := db.GetDB()
	var user models.User

	result := db.Where("id = ?", req.Id).First(&user)

	if result.Error != nil {
		return nil, result.Error
	}

	db.Delete(&user)

	return &pb.DeleteUserResponse{
		User: user.ToGRPC(),
	}, nil
}

func RegisterUserService(s *grpc.Server) {
	pb.RegisterUserServiceServer(s, &userSevice{})
}
