package grpcserver

import (
	"context"
	"errors"
	"gkeeper/api/proto"
	"gkeeper/internal/storage"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (gs *GKeeperServer) Register(ctx context.Context, req *proto.RegisterRequest) (*proto.RegisterResponse, error) {
	var response proto.RegisterResponse

	hashedPassword, hashErr := hashPassword(req.GetPassword())
	if hashErr != nil {
		return &response, status.Errorf(codes.Unauthenticated, "invalid password: %v", hashErr)
	}

	_, err := gs.storage.CreateUser(ctx, req.GetEmail(), hashedPassword)
	if err != nil {
		if errors.Is(err, storage.ErrUserAlreadyExists) {
			return &response, status.Errorf(codes.AlreadyExists, "user already exists")
		}
		return &response, status.Errorf(codes.Internal, "internal error: %v", err.Error())
	}

	response.SetResult("User registered successfully")

	return &response, nil
}

func (gs *GKeeperServer) Login(ctx context.Context, req *proto.LoginRequest) (*proto.LoginResponse, error) {
	var response proto.LoginResponse

	user, err := gs.storage.GetUserByEmail(ctx, req.GetEmail())
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return &response, status.Errorf(codes.Unauthenticated, "user not found")
		}
		return &response, status.Errorf(codes.Internal, "intrnal error")
	}

	if !checkPasswordHash(req.GetPassword(), user.PasswordHash) {
		return &response, status.Errorf(codes.Unauthenticated, "user not found")
	}

	token, err := gs.jwtManager.GenerateToken(user.ID, user.Email)
	if err != nil {
		gs.logger.Error("failed to generate token", zap.Error(err))
		return &response, status.Errorf(codes.Internal, "failed to generate token")
	}

	response.SetResult(token)

	return &response, nil
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
