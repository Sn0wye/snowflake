package grpc

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/getsnowflake/snowflake/helium/pkg/jwt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func newTestJWT(t *testing.T) *jwt.JWT {
	t.Helper()
	jwter, err := jwt.NewTestJWT("test-access-key-32byteslong!!", "test-refresh-key-32bytes!!", "helium-test")
	if err != nil {
		t.Fatalf("failed to create test JWT: %v", err)
	}
	return jwter
}

func TestRecoveryInterceptor_NoPanic(t *testing.T) {
	core, _ := observer.New(zapcore.InfoLevel)
	log := zap.New(core)
	interceptor := RecoveryInterceptor(log)

	handler := func(ctx context.Context, req any) (any, error) {
		return "ok", nil
	}

	resp, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: "/test/Method"}, handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "ok" {
		t.Fatalf("expected 'ok', got %v", resp)
	}
}

func TestRecoveryInterceptor_Panic(t *testing.T) {
	core, logs := observer.New(zapcore.ErrorLevel)
	log := zap.New(core)
	interceptor := RecoveryInterceptor(log)

	handler := func(ctx context.Context, req any) (any, error) {
		panic("boom")
	}

	_, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: "/test/PanickingMethod"}, handler)
	if err == nil {
		t.Fatal("expected error from panic recovery")
	}
	st, _ := status.FromError(err)
	if st.Code() != codes.Internal {
		t.Fatalf("expected Internal, got %v", st.Code())
	}
	if !strings.Contains(st.Message(), "internal server error") {
		t.Fatalf("expected internal server error message, got %s", st.Message())
	}

	if logs.Len() == 0 {
		t.Fatal("expected error log entry")
	}
	entry := logs.All()[0]
	if entry.Level != zapcore.ErrorLevel {
		t.Fatalf("expected ErrorLevel, got %v", entry.Level)
	}
	if entry.ContextMap()["method"] != "/test/PanickingMethod" {
		t.Fatalf("expected method in log, got %v", entry.ContextMap()["method"])
	}
}

func TestLoggingInterceptor_Success(t *testing.T) {
	core, logs := observer.New(zapcore.InfoLevel)
	log := zap.New(core)
	interceptor := LoggingInterceptor(log)

	handler := func(ctx context.Context, req any) (any, error) {
		return "ok", nil
	}

	resp, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: "/test/Method"}, handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "ok" {
		t.Fatalf("expected 'ok'")
	}

	entries := logs.FilterLevelExact(zapcore.InfoLevel)
	if entries.Len() == 0 {
		t.Fatal("expected info log entry")
	}
	entry := entries.All()[0]
	ctxMap := entry.ContextMap()
	if ctxMap["method"] != "/test/Method" {
		t.Fatalf("expected method /test/Method, got %v", ctxMap["method"])
	}
	if ctxMap["code"] != "OK" {
		t.Fatalf("expected code OK, got %v", ctxMap["code"])
	}
	if _, ok := ctxMap["duration"]; !ok {
		t.Fatal("expected duration field")
	}
}

func TestLoggingInterceptor_Error(t *testing.T) {
	core, logs := observer.New(zapcore.WarnLevel)
	log := zap.New(core)
	interceptor := LoggingInterceptor(log)

	handler := func(ctx context.Context, req any) (any, error) {
		return nil, status.Error(codes.NotFound, "not found")
	}

	_, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: "/test/NotFound"}, handler)
	if err == nil {
		t.Fatal("expected error")
	}

	entries := logs.FilterLevelExact(zapcore.WarnLevel)
	if entries.Len() == 0 {
		t.Fatal("expected warn log entry")
	}
	entry := entries.All()[0]
	ctxMap := entry.ContextMap()
	if ctxMap["code"] != "NotFound" {
		t.Fatalf("expected code NotFound, got %v", ctxMap["code"])
	}
}

func TestAuthInterceptor_ValidToken(t *testing.T) {
	jwter := newTestJWT(t)
	token := genToken(t, jwter, "user-1", time.Now().Add(time.Hour))
	interceptor := AuthInterceptor(jwter)

	md := metadata.Pairs("authorization", "Bearer "+token)
	ctx := metadata.NewIncomingContext(context.Background(), md)

	handler := func(ctx context.Context, req any) (any, error) {
		claims := ClaimsFromContext(ctx)
		if claims == nil {
			return nil, status.Error(codes.Internal, "claims not in context")
		}
		if claims.Subject != "user-1" {
			return nil, status.Errorf(codes.Internal, "expected subject user-1, got %s", claims.Subject)
		}
		return "ok", nil
	}

	resp, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/pb.UserService/GetUsers"}, handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "ok" {
		t.Fatalf("expected 'ok', got %v", resp)
	}
}

func TestAuthInterceptor_MissingMetadata(t *testing.T) {
	jwter := newTestJWT(t)
	interceptor := AuthInterceptor(jwter)

	_, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: "/pb.UserService/GetUsers"}, nil)
	if err == nil {
		t.Fatal("expected error for missing metadata")
	}
	st, _ := status.FromError(err)
	if st.Code() != codes.Unauthenticated {
		t.Fatalf("expected Unauthenticated, got %v", st.Code())
	}
}

func TestAuthInterceptor_MissingToken(t *testing.T) {
	jwter := newTestJWT(t)
	interceptor := AuthInterceptor(jwter)

	md := metadata.Pairs("x-custom", "value")
	ctx := metadata.NewIncomingContext(context.Background(), md)

	_, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/pb.UserService/GetUsers"}, nil)
	if err == nil {
		t.Fatal("expected error for missing token")
	}
	st, _ := status.FromError(err)
	if st.Code() != codes.Unauthenticated {
		t.Fatalf("expected Unauthenticated, got %v", st.Code())
	}
}

func TestAuthInterceptor_InvalidToken(t *testing.T) {
	jwter := newTestJWT(t)
	interceptor := AuthInterceptor(jwter)

	md := metadata.Pairs("authorization", "Bearer garbage")
	ctx := metadata.NewIncomingContext(context.Background(), md)

	_, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/pb.UserService/GetUsers"}, nil)
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
	st, _ := status.FromError(err)
	if st.Code() != codes.Unauthenticated {
		t.Fatalf("expected Unauthenticated, got %v", st.Code())
	}
}

func TestAuthInterceptor_ExpiredToken(t *testing.T) {
	jwter := newTestJWT(t)
	token := genToken(t, jwter, "user-1", time.Now().Add(-time.Hour))
	interceptor := AuthInterceptor(jwter)

	md := metadata.Pairs("authorization", "Bearer "+token)
	ctx := metadata.NewIncomingContext(context.Background(), md)

	_, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/pb.UserService/GetUsers"}, nil)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
	st, _ := status.FromError(err)
	if st.Code() != codes.Unauthenticated {
		t.Fatalf("expected Unauthenticated, got %v", st.Code())
	}
}

func TestAuthInterceptor_TokenWithoutBearer(t *testing.T) {
	jwter := newTestJWT(t)
	token := genToken(t, jwter, "user-1", time.Now().Add(time.Hour))
	interceptor := AuthInterceptor(jwter)

	md := metadata.Pairs("authorization", token)
	ctx := metadata.NewIncomingContext(context.Background(), md)

	handler := func(ctx context.Context, req any) (any, error) {
		claims := ClaimsFromContext(ctx)
		if claims == nil {
			return nil, status.Error(codes.Internal, "claims not in context")
		}
		return "ok", nil
	}

	resp, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/pb.UserService/GetUsers"}, handler)
	if err != nil {
		t.Fatalf("token without Bearer prefix should work: %v", err)
	}
	if resp != "ok" {
		t.Fatalf("expected 'ok'")
	}
}

func TestAuthInterceptor_SkippedMethods(t *testing.T) {
	jwter := newTestJWT(t)
	interceptor := AuthInterceptor(jwter)

	skipped := []string{"/pb.AuthService/ValidateToken", "/pb.AuthService/ParseToken"}

	for _, method := range skipped {
		t.Run(method, func(t *testing.T) {
			handler := func(ctx context.Context, req any) (any, error) {
				return "ok", nil
			}
			resp, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: method}, handler)
			if err != nil {
				t.Fatalf("expected no error for skipped method %s: %v", method, err)
			}
			if resp != "ok" {
				t.Fatalf("expected 'ok' for skipped method %s", method)
			}
		})
	}
}
