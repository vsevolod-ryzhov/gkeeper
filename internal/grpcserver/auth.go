package grpcserver

import (
	"context"

	"gkeeper/api/proto"
)

func (gs *GKeeperServer) Register(ctx context.Context, req *proto.RegisterRequest) (*proto.RegisterResponse, error) {
	var response *proto.RegisterResponse

	return response, nil
}

func (gs *GKeeperServer) Login(ctx context.Context, req *proto.LoginRequest) (*proto.LoginResponse, error) {
	var response *proto.LoginResponse

	return response, nil
}
