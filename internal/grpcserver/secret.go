package grpcserver

import (
	"context"
	"fmt"
	"gkeeper/api/proto"
)

func (gs *GKeeperServer) CreateSecret(ctx context.Context, req *proto.CreateSecretRequest) (*proto.CreateSecretResponse, error) {
	var response proto.CreateSecretResponse

	fmt.Println(req)

	return &response, nil
}

func (gs *GKeeperServer) UpdateSecret(ctx context.Context, req *proto.UpdateSecretRequest) (*proto.UpdateSecretResponse, error) {
	var response proto.UpdateSecretResponse

	return &response, nil
}

func (gs *GKeeperServer) DeleteSecret(ctx context.Context, req *proto.DeleteSecretRequest) (*proto.DeleteSecretResponse, error) {
	var response proto.DeleteSecretResponse

	return &response, nil
}

func (gs *GKeeperServer) GetSecrets(ctx context.Context, req *proto.GetSecretsRequest) (*proto.GetSecretsResponse, error) {
	var response proto.GetSecretsResponse

	return &response, nil
}

func (gs *GKeeperServer) GetSecret(ctx context.Context, req *proto.GetSecretRequest) (*proto.GetSecretResponse, error) {
	var response proto.GetSecretResponse

	return &response, nil
}
