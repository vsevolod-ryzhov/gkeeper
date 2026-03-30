package grpcserver

import (
	"context"
	"fmt"
	pb "gkeeper/api/proto"
	"gkeeper/internal/model"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

const maxBinarySecretSize = 50 * 1024 * 102

func (gs *GKeeperServer) CreateSecret(ctx context.Context, req *pb.CreateSecretRequest) (*pb.CreateSecretResponse, error) {
	userID, ok := ctx.Value(ctxKeyUserID).(uuid.UUID)
	if !ok || userID == uuid.Nil {
		return nil, status.Errorf(codes.Unauthenticated, "user not authenticated")
	}

	secretType := model.ProtoToSecretType(req.GetType())
	encryptedData := string(req.GetEncryptedData())

	if secretType == model.SecretTypeBinary {
		if len(req.GetEncryptedData()) > maxBinarySecretSize {
			return nil, status.Errorf(codes.InvalidArgument, "file size exceeds maximum of 50MB")
		}

		// Create DB record first with empty data to get the ID
		secret, err := gs.storage.CreateSecret(ctx, userID.String(), req.GetTitle(), secretType, "", req.GetMetadata(), req.GetFilePath())
		if err != nil {
			gs.logger.Error("failed to create secret", zap.Error(err))
			return nil, status.Errorf(codes.Internal, "failed to create secret")
		}

		objectKey := fmt.Sprintf("%s/%s", userID.String(), secret.ID.String())
		if err := gs.fileStorage.Upload(ctx, objectKey, req.GetEncryptedData()); err != nil {
			gs.logger.Error("failed to upload binary to storage", zap.Error(err))
			return nil, status.Errorf(codes.Internal, "failed to upload file")
		}

		// Update the record with the object key
		secret, err = gs.storage.UpdateSecret(ctx, userID.String(), secret.ID.String(), req.GetTitle(), objectKey, req.GetMetadata(), req.GetFilePath())
		if err != nil {
			gs.logger.Error("failed to update secret with object key", zap.Error(err))
			return nil, status.Errorf(codes.Internal, "failed to create secret")
		}

		gs.logger.Info("binary secret created", zap.String("id", secret.ID.String()), zap.String("user_id", userID.String()))

		return pb.CreateSecretResponse_builder{
			Id:        proto.String(secret.ID.String()),
			CreatedAt: proto.String(secret.CreatedAt.Format("2006-01-02T15:04:05Z07:00")),
		}.Build(), nil
	}

	secret, err := gs.storage.CreateSecret(ctx, userID.String(), req.GetTitle(), secretType, encryptedData, req.GetMetadata(), req.GetFilePath())
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
	userID, ok := ctx.Value(ctxKeyUserID).(uuid.UUID)
	if !ok || userID == uuid.Nil {
		return nil, status.Errorf(codes.Unauthenticated, "user not authenticated")
	}

	secretType := model.ProtoToSecretType(req.GetType())
	encryptedData := string(req.GetEncryptedData())

	if secretType == model.SecretTypeBinary && len(req.GetEncryptedData()) > 0 {
		if len(req.GetEncryptedData()) > maxBinarySecretSize {
			return nil, status.Errorf(codes.InvalidArgument, "file size exceeds maximum of 50MB")
		}

		objectKey := fmt.Sprintf("%s/%s", userID.String(), req.GetId())
		if err := gs.fileStorage.Upload(ctx, objectKey, req.GetEncryptedData()); err != nil {
			gs.logger.Error("failed to upload binary to storage", zap.Error(err))
			return nil, status.Errorf(codes.Internal, "failed to upload file")
		}
		encryptedData = objectKey
	}

	secret, err := gs.storage.UpdateSecret(ctx, userID.String(), req.GetId(), req.GetTitle(), encryptedData, req.GetMetadata(), req.GetFilePath())
	if err != nil {
		gs.logger.Error("failed to update secret", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to update secret")
	}

	return pb.UpdateSecretResponse_builder{
		UpdatedAt: proto.String(secret.UpdatedAt.Format("2006-01-02T15:04:05Z07:00")),
	}.Build(), nil
}

func (gs *GKeeperServer) DeleteSecret(ctx context.Context, req *pb.DeleteSecretRequest) (*pb.DeleteSecretResponse, error) {
	userID, ok := ctx.Value(ctxKeyUserID).(uuid.UUID)
	if !ok || userID == uuid.Nil {
		return nil, status.Errorf(codes.Unauthenticated, "user not authenticated")
	}

	err := gs.storage.DeleteSecret(ctx, userID.String(), req.GetId())
	if err != nil {
		gs.logger.Error("failed to delete secret", zap.Error(err), zap.String("id", req.GetId()))
		return nil, status.Errorf(codes.Internal, "failed to delete secret")
	}

	gs.logger.Info("secret deleted", zap.String("id", req.GetId()), zap.String("user_id", userID.String()))

	return pb.DeleteSecretResponse_builder{
		Success: proto.Bool(true),
	}.Build(), nil
}

func (gs *GKeeperServer) GetSecrets(ctx context.Context, req *pb.GetSecretsRequest) (*pb.GetSecretsResponse, error) {
	userID, ok := ctx.Value(ctxKeyUserID).(uuid.UUID)
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

		// Don't include binary data in list responses — client fetches via GetSecret
		var encryptedData []byte
		if s.Type != model.SecretTypeBinary {
			encryptedData = []byte(s.EncryptedData)
		}

		secretType := model.SecretTypeToProto(s.Type)
		pbSecrets = append(pbSecrets, pb.Secret_builder{
			Id:            proto.String(s.ID.String()),
			UserId:        proto.String(s.UserID.String()),
			Title:         proto.String(s.Title),
			Type:          &secretType,
			EncryptedData: encryptedData,
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
	userID, ok := ctx.Value(ctxKeyUserID).(uuid.UUID)
	if !ok || userID == uuid.Nil {
		return nil, status.Errorf(codes.Unauthenticated, "user not authenticated")
	}

	secret, err := gs.storage.GetSecretByID(ctx, userID.String(), req.GetId())
	if err != nil {
		gs.logger.Error("failed to get secret", zap.Error(err), zap.String("id", req.GetId()))
		return nil, status.Errorf(codes.NotFound, "secret not found")
	}

	encryptedData := []byte(secret.EncryptedData)

	// For binary secrets, fetch the actual data from object storage
	if secret.Type == model.SecretTypeBinary && secret.EncryptedData != "" {
		data, err := gs.fileStorage.Download(ctx, secret.EncryptedData)
		if err != nil {
			gs.logger.Error("failed to download binary from storage", zap.Error(err))
			return nil, status.Errorf(codes.Internal, "failed to retrieve file")
		}
		encryptedData = data
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

	secretType := model.SecretTypeToProto(secret.Type)
	return pb.GetSecretResponse_builder{
		Secret: pb.Secret_builder{
			Id:            proto.String(secret.ID.String()),
			UserId:        proto.String(secret.UserID.String()),
			Title:         proto.String(secret.Title),
			Type:          &secretType,
			EncryptedData: encryptedData,
			Metadata:      proto.String(string(secret.Metadata)),
			FilePath:      filePath,
			CreatedAt:     proto.String(secret.CreatedAt.Format("2006-01-02T15:04:05Z07:00")),
			UpdatedAt:     proto.String(secret.UpdatedAt.Format("2006-01-02T15:04:05Z07:00")),
			DeletedAt:     deletedAt,
		}.Build(),
	}.Build(), nil
}
