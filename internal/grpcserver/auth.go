package grpcserver

import (
	"context"
	"fmt"

	"gkeeper/api/proto"
)

func (gs *GKeeperServer) Register(ctx context.Context, req *proto.RegisterRequest) (*proto.RegisterResponse, error) {
	var response proto.RegisterResponse

	return &response, nil
}

func (gs *GKeeperServer) Login(ctx context.Context, req *proto.LoginRequest) (*proto.LoginResponse, error) {
	var response proto.LoginResponse

	// TODO: real implementation here
	response.SetResult(fmt.Sprintf("Login %s legged in", req.GetEmail()))

	return &response, nil
}
