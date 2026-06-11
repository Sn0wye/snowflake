package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/getsnowflake/snowflake/helium/pb"
	"github.com/getsnowflake/snowflake/helium/pkg/jwt"
	"github.com/getsnowflake/snowflake/helium/src/models"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupUserTest(t *testing.T) (*userService, *jwt.JWT) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	jwter, err := jwt.NewTestJWT("test-access-key-32byteslong!!", "test-refresh-key-32bytes!!", "helium-test")
	if err != nil {
		t.Fatalf("failed to create test JWT: %v", err)
	}
	return &userService{db: db, jwter: jwter}, jwter
}

func seedUser(t *testing.T, svc *userService, id, name, username, email string) {
	t.Helper()
	user := models.User{
		Name:     name,
		Username: username,
		Email:    email,
		Password: "hashed-password",
	}
	if err := svc.db.Create(&user).Error; err != nil {
		t.Fatalf("failed to seed user: %v", err)
	}
	if id != "" {
		svc.db.Exec("UPDATE users SET id = ? WHERE email = ?", id, email)
	}
}

func TestGetUsers_Empty(t *testing.T) {
	svc, _ := setupUserTest(t)
	resp, err := svc.GetUsers(context.Background(), &emptypb.Empty{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Users) != 0 {
		t.Fatalf("expected 0 users, got %d", len(resp.Users))
	}
}

func TestGetUsers_ReturnsAll(t *testing.T) {
	svc, _ := setupUserTest(t)
	seedUser(t, svc, "", "Alice", "alice", "alice@test.com")
	seedUser(t, svc, "", "Bob", "bob", "bob@test.com")
	resp, err := svc.GetUsers(context.Background(), &emptypb.Empty{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(resp.Users))
	}
	names := map[string]bool{}
	for _, u := range resp.Users {
		names[u.Name] = true
	}
	if !names["Alice"] || !names["Bob"] {
		t.Fatalf("expected Alice and Bob in response, got %v", names)
	}
}

func TestGetUser_Found(t *testing.T) {
	svc, _ := setupUserTest(t)
	seedUser(t, svc, "1b7e4a5e-9d3c-4e2f-8a1d-6c5b9e0f1a2b", "Alice", "alice", "alice@test.com")
	resp, err := svc.GetUser(context.Background(), &pb.GetUserRequest{Id: "1b7e4a5e-9d3c-4e2f-8a1d-6c5b9e0f1a2b"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.User.Name != "Alice" {
		t.Fatalf("expected Alice, got %s", resp.User.Name)
	}
	if resp.User.Email != "alice@test.com" {
		t.Fatalf("expected alice@test.com, got %s", resp.User.Email)
	}
}

func TestGetUser_NotFound(t *testing.T) {
	svc, _ := setupUserTest(t)
	_, err := svc.GetUser(context.Background(), &pb.GetUserRequest{Id: "nonexistent-id"})
	if err == nil {
		t.Fatal("expected error for non-existent user")
	}
}

func TestDeleteUser_Success(t *testing.T) {
	svc, jwter := setupUserTest(t)
	userID := "1b7e4a5e-9d3c-4e2f-8a1d-6c5b9e0f1a2b"
	seedUser(t, svc, userID, "Alice", "alice", "alice@test.com")
	token := genToken(t, jwter, userID, time.Now().Add(time.Hour))
	md := metadata.Pairs("authorization", "Bearer "+token)
	ctx := metadata.NewIncomingContext(context.Background(), md)
	resp, err := svc.DeleteUser(ctx, &pb.DeleteUserRequest{Id: userID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.User.Name != "Alice" {
		t.Fatalf("expected deleted user Alice, got %s", resp.User.Name)
	}
	_, err = svc.GetUser(context.Background(), &pb.GetUserRequest{Id: userID})
	if err == nil {
		t.Fatal("expected user to be deleted")
	}
}

func TestDeleteUser_MissingMetadata(t *testing.T) {
	svc, _ := setupUserTest(t)
	_, err := svc.DeleteUser(context.Background(), &pb.DeleteUserRequest{Id: "some-id"})
	if err == nil {
		t.Fatal("expected error for missing metadata")
	}
	st, _ := status.FromError(err)
	if st.Code() != codes.Unauthenticated {
		t.Fatalf("expected Unauthenticated, got %v", st.Code())
	}
}

func TestDeleteUser_MissingAuthorizationHeader(t *testing.T) {
	svc, _ := setupUserTest(t)
	md := metadata.Pairs("x-custom", "value")
	ctx := metadata.NewIncomingContext(context.Background(), md)
	_, err := svc.DeleteUser(ctx, &pb.DeleteUserRequest{Id: "some-id"})
	if err == nil {
		t.Fatal("expected error for missing authorization")
	}
	st, _ := status.FromError(err)
	if st.Code() != codes.Unauthenticated {
		t.Fatalf("expected Unauthenticated, got %v", st.Code())
	}
}

func TestDeleteUser_InvalidToken(t *testing.T) {
	svc, _ := setupUserTest(t)
	md := metadata.Pairs("authorization", "Bearer garbage")
	ctx := metadata.NewIncomingContext(context.Background(), md)
	_, err := svc.DeleteUser(ctx, &pb.DeleteUserRequest{Id: "some-id"})
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
	st, _ := status.FromError(err)
	if st.Code() != codes.Unauthenticated {
		t.Fatalf("expected Unauthenticated, got %v", st.Code())
	}
}

func TestDeleteUser_CannotDeleteOtherUser(t *testing.T) {
	svc, jwter := setupUserTest(t)
	userA := "1b7e4a5e-9d3c-4e2f-8a1d-6c5b9e0f1a2b"
	userB := "b2c3d4e5-f6a7-8901-bcde-f12345678901"
	seedUser(t, svc, userA, "Alice", "alice", "alice@test.com")
	seedUser(t, svc, userB, "Bob", "bob", "bob@test.com")
	token := genToken(t, jwter, userA, time.Now().Add(time.Hour))
	md := metadata.Pairs("authorization", "Bearer "+token)
	ctx := metadata.NewIncomingContext(context.Background(), md)
	_, err := svc.DeleteUser(ctx, &pb.DeleteUserRequest{Id: userB})
	if err == nil {
		t.Fatal("expected error when deleting another user")
	}
	st, _ := status.FromError(err)
	if st.Code() != codes.PermissionDenied {
		t.Fatalf("expected PermissionDenied, got %v", st.Code())
	}
}

func TestDeleteUser_NotFound(t *testing.T) {
	svc, jwter := setupUserTest(t)
	userID := "1b7e4a5e-9d3c-4e2f-8a1d-6c5b9e0f1a2b"
	token := genToken(t, jwter, userID, time.Now().Add(time.Hour))
	md := metadata.Pairs("authorization", "Bearer "+token)
	ctx := metadata.NewIncomingContext(context.Background(), md)
	_, err := svc.DeleteUser(ctx, &pb.DeleteUserRequest{Id: userID})
	if err == nil {
		t.Fatal("expected error for non-existent user")
	}
}

func TestGetUsers_UserFieldsMapped(t *testing.T) {
	svc, _ := setupUserTest(t)
	user := models.User{
		Name:         "Alice",
		Username:     "alice",
		Email:        "alice@test.com",
		Password:     "hashed-password",
		AnnualIncome: 100000,
		Debt:         50000,
		AssetsValue:  200000,
	}
	if err := svc.db.Create(&user).Error; err != nil {
		t.Fatalf("failed to seed user: %v", err)
	}
	resp, _ := svc.GetUsers(context.Background(), &emptypb.Empty{})
	u := resp.Users[0]
	if u.AnnualIncome != 100000 {
		t.Fatalf("expected AnnualIncome 100000, got %d", u.AnnualIncome)
	}
	if u.Debt != 50000 {
		t.Fatalf("expected Debt 50000, got %d", u.Debt)
	}
	if u.AssetsValue != 200000 {
		t.Fatalf("expected AssetsValue 200000, got %d", u.AssetsValue)
	}
	if u.Id == "" {
		t.Fatal("expected Id to be set")
	}
	if u.CreatedAt == "" {
		t.Fatal("expected CreatedAt to be set")
	}
	if u.UpdatedAt == "" {
		t.Fatal("expected UpdatedAt to be set")
	}
}

func TestDeleteUser_TokenWithoutBearer(t *testing.T) {
	svc, jwter := setupUserTest(t)
	userID := "1b7e4a5e-9d3c-4e2f-8a1d-6c5b9e0f1a2b"
	seedUser(t, svc, userID, "Alice", "alice", "alice@test.com")
	token := genToken(t, jwter, userID, time.Now().Add(time.Hour))
	md := metadata.Pairs("authorization", token)
	ctx := metadata.NewIncomingContext(context.Background(), md)
	resp, err := svc.DeleteUser(ctx, &pb.DeleteUserRequest{Id: userID})
	if err != nil {
		t.Fatalf("token without Bearer prefix should work: %v", err)
	}
	if resp.User.Name != "Alice" {
		t.Fatalf("expected Alice, got %s", resp.User.Name)
	}
}

func TestGetUsers_ResponseFormat(t *testing.T) {
	svc, _ := setupUserTest(t)
	seedUser(t, svc, "", "Test", "test", "test@test.com")
	resp, err := svc.GetUsers(context.Background(), &emptypb.Empty{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Users == nil {
		t.Fatal("Users should not be nil, should be empty slice")
	}
	if len(resp.Users) != 1 {
		t.Fatalf("expected 1 user, got %d", len(resp.Users))
	}
}
