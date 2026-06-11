package grpc

import (
	"context"
	"runtime/debug"
	"time"

	"github.com/getsnowflake/snowflake/helium/pkg/config"
	"github.com/getsnowflake/snowflake/helium/pkg/jwt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type claimsCtxKey struct{}

var authSkippedMethods = map[string]bool{
	"/pb.AuthService/ValidateToken": true,
	"/pb.AuthService/ParseToken":    true,
}

func RecoveryInterceptor(log *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		defer func() {
			if r := recover(); r != nil {
				log.Error("gRPC panic recovered",
					zap.String("method", info.FullMethod),
					zap.Any("panic", r),
					zap.String("stack", string(debug.Stack())),
				)
				err = status.Errorf(codes.Internal, "internal server error")
			}
		}()
		return handler(ctx, req)
	}
}

func LoggingInterceptor(log *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start)

		fields := []zapcore.Field{
			zap.String("method", info.FullMethod),
			zap.Duration("duration", duration),
		}

		if err != nil {
			st, _ := status.FromError(err)
			fields = append(fields, zap.String("code", st.Code().String()), zap.Error(err))
			log.Warn("gRPC call failed", fields...)
		} else {
			fields = append(fields, zap.String("code", codes.OK.String()))
			log.Info("gRPC call completed", fields...)
		}

		return resp, err
	}
}

func AuthInterceptor(jwter *jwt.JWT) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if authSkippedMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		vals := md.Get("authorization")
		if len(vals) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing authorization token")
		}

		claims, err := jwter.ParseToken(vals[0])
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
		}

		ctx = context.WithValue(ctx, claimsCtxKey{}, claims)
		return handler(ctx, req)
	}
}

func ClaimsFromContext(ctx context.Context) *jwt.Claims {
	claims, _ := ctx.Value(claimsCtxKey{}).(*jwt.Claims)
	return claims
}

func newJWT() *jwt.JWT {
	return jwt.NewJwt(config.GetConfig())
}
