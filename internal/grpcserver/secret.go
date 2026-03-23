package grpcserver

import (
	"context"
	pb "gkeeper/api/proto"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func (gs *GKeeperServer) CreateSecret(ctx context.Context, req *pb.CreateSecretRequest) (*pb.CreateSecretResponse, error) {
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

	return pb.CreateSecretResponse_builder{
		Id:        proto.String(secret.ID.String()),
		CreatedAt: proto.String(secret.CreatedAt.Format("2006-01-02T15:04:05Z07:00")),
	}.Build(), nil
}

func (gs *GKeeperServer) UpdateSecret(ctx context.Context, req *pb.UpdateSecretRequest) (*pb.UpdateSecretResponse, error) {
	var response pb.UpdateSecretResponse

	return &response, nil
}

func (gs *GKeeperServer) DeleteSecret(ctx context.Context, req *pb.DeleteSecretRequest) (*pb.DeleteSecretResponse, error) {
	var response pb.DeleteSecretResponse

	return &response, nil
}

func (gs *GKeeperServer) GetSecrets(ctx context.Context, req *pb.GetSecretsRequest) (*pb.GetSecretsResponse, error) {
	userID, ok := ctx.Value("user_id").(uuid.UUID)
	if !ok || userID == uuid.Nil {
		return nil, status.Errorf(codes.Unauthenticated, "user not authenticated")
	}

	secrets, err := gs.storage.GetSecretsByUserID(ctx, userID.String())
	if err != nil {
		gs.logger.Error("failed to get secrets", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get secrets")
	}

	pbSecrets := make([]*pb.Secret, 0, len(secrets))
	for _, s := range secrets {
		var deletedAt *string
		if s.DeletedAt != nil {
			v := s.DeletedAt.Format("2006-01-02T15:04:05Z07:00")
			deletedAt = &v
		}
		var filePath *string
		if s.FilePath != nil {
			filePath = s.FilePath
		}

		pbSecrets = append(pbSecrets, pb.Secret_builder{
			Id:            proto.String(s.ID.String()),
			UserId:        proto.String(s.UserID.String()),
			Title:         proto.String(s.Title),
			Type:          proto.String(s.Type),
			EncryptedData: []byte(s.EncryptedData),
			Metadata:      proto.String(string(s.Metadata)),
			FilePath:      filePath,
			CreatedAt:     proto.String(s.CreatedAt.Format("2006-01-02T15:04:05Z07:00")),
			UpdatedAt:     proto.String(s.UpdatedAt.Format("2006-01-02T15:04:05Z07:00")),
			DeletedAt:     deletedAt,
		}.Build())
	}

	return pb.GetSecretsResponse_builder{
		Secrets: pbSecrets,
	}.Build(), nil
}

func (gs *GKeeperServer) GetSecret(ctx context.Context, req *pb.GetSecretRequest) (*pb.GetSecretResponse, error) {
	userID, ok := ctx.Value("user_id").(uuid.UUID)
	if !ok || userID == uuid.Nil {
		return nil, status.Errorf(codes.Unauthenticated, "user not authenticated")
	}

	secret, err := gs.storage.GetSecretByID(ctx, userID.String(), req.GetId())
	if err != nil {
		gs.logger.Error("failed to get secret", zap.Error(err), zap.String("id", req.GetId()))
		return nil, status.Errorf(codes.NotFound, "secret not found")
	}

	var deletedAt *string
	if secret.DeletedAt != nil {
		v := secret.DeletedAt.Format("2006-01-02T15:04:05Z07:00")
		deletedAt = &v
	}
	var filePath *string
	if secret.FilePath != nil {
		filePath = secret.FilePath
	}

	return pb.GetSecretResponse_builder{
		Secret: pb.Secret_builder{
			Id:            proto.String(secret.ID.String()),
			UserId:        proto.String(secret.UserID.String()),
			Title:         proto.String(secret.Title),
			Type:          proto.String(secret.Type),
			EncryptedData: []byte(secret.EncryptedData),
			Metadata:      proto.String(string(secret.Metadata)),
			FilePath:      filePath,
			CreatedAt:     proto.String(secret.CreatedAt.Format("2006-01-02T15:04:05Z07:00")),
			UpdatedAt:     proto.String(secret.UpdatedAt.Format("2006-01-02T15:04:05Z07:00")),
			DeletedAt:     deletedAt,
		}.Build(),
	}.Build(), nil
}
