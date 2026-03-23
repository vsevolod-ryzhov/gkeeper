package grpcserver

import (
	"context"
	"gkeeper/api/proto"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	protolib "google.golang.org/protobuf/proto"
)

func (gs *GKeeperServer) CreateSecret(ctx context.Context, req *proto.CreateSecretRequest) (*proto.CreateSecretResponse, error) {
	userID, ok := ctx.Value("user_id").(uuid.UUID)
	if !ok || userID == uuid.Nil {
		return nil, status.Errorf(codes.Unauthenticated, "user not authenticated")
	}

	secret, err := gs.storage.CreateSecret(
		ctx,
		userID.String(),
		req.GetTitle(),
		req.GetType(),
		string(req.GetEncryptedData()),
		req.GetMetadata(),
		req.GetFilePath(),
	)
	if err != nil {
		gs.logger.Error("failed to create secret", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to create secret")
	}

	gs.logger.Info("secret created", zap.String("id", secret.ID.String()), zap.String("user_id", userID.String()))

	return proto.CreateSecretResponse_builder{
		Id:        protolib.String(secret.ID.String()),
		CreatedAt: protolib.String(secret.CreatedAt.Format("2006-01-02T15:04:05Z07:00")),
	}.Build(), nil
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
