package grpcserver

import (
	"context"
	"errors"
	"fmt"
	"gkeeper/internal/jwt"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type contextKey string

const (
	ctxKeyUserID contextKey = "user_id"
	ctxKeyEmail  contextKey = "email"
)

type AuthInterceptor struct {
	jwtManager *jwt.JWTManager
}

func NewAuthInterceptor(jwtManager *jwt.JWTManager) *AuthInterceptor {
	return &AuthInterceptor{
		jwtManager: jwtManager,
	}
}

func (interceptor *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if !interceptor.isProtectedMethod(info.FullMethod) {
			return handler(ctx, req)
		}

		claims, err := interceptor.authorize(ctx)
		if err != nil {
			return nil, err
		}

		ctx = context.WithValue(ctx, ctxKeyUserID, claims.UserID)
		ctx = context.WithValue(ctx, ctxKeyEmail, claims.Email)

		return handler(ctx, req)
	}
}

func (interceptor *AuthInterceptor) isProtectedMethod(method string) bool {
	publicMethods := map[string]bool{
		"/vsevolodryzhov.gkeeper.proto.GKeeper/Register": true,
		"/vsevolodryzhov.gkeeper.proto.GKeeper/Login":    true,
	}

	return !publicMethods[method]
}

func (interceptor *AuthInterceptor) authorize(ctx context.Context) (*jwt.Claims, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "metadata is not provided")
	}
	fmt.Println(md)
	values := md["authorization"]
	if len(values) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "authorization token is not provided")
	}

	token := values[0]

	parts := strings.Split(token, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return nil, status.Errorf(codes.Unauthenticated, "invalid authorization header format")
	}

	claims, err := interceptor.jwtManager.VerifyToken(parts[1])
	if err != nil {
		if errors.Is(err, jwt.ErrExpiredToken) {
			return nil, status.Errorf(codes.Unauthenticated, "token is expired")
		}
		return nil, status.Errorf(codes.Unauthenticated, "invalid token")
	}

	return claims, nil
}
