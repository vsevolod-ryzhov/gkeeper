package grpcserver

import (
	"context"
	"fmt"
	"gkeeper/internal/storage"
	"strconv"

	"errors"
	"gkeeper/api/proto"

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

	user, err := gs.storage.CreateUser(ctx, req.GetEmail(), hashedPassword)
	if err != nil {
		if errors.Is(err, storage.ErrUserAlreadyExists) {
			return &response, status.Errorf(codes.AlreadyExists, "user already exists")
		}
		return &response, status.Errorf(codes.Internal, "internal error: %v", err.Error())
	}

	response.SetResult(strconv.FormatInt(user.ID, 10))

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

	// TODO: real implementation here
	response.SetResult(fmt.Sprintf("Login %s legged in", req.GetEmail()))

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
