package grpcserver

import (
	"context"
	"gkeeper/internal/jwt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func newTestInterceptor() (*AuthInterceptor, *jwt.JWTManager) {
	jwtManager := jwt.NewJWTManager("test-secret", 1*time.Hour)
	return NewAuthInterceptor(jwtManager), jwtManager
}

func TestAuthInterceptor_PublicMethod(t *testing.T) {
	interceptor, _ := newTestInterceptor()

	handlerCalled := false
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		handlerCalled = true
		return "ok", nil
	}

	unary := interceptor.Unary()
	resp, err := unary(
		context.Background(),
		nil,
		&grpc.UnaryServerInfo{FullMethod: "/vsevolodryzhov.gkeeper.proto.GKeeper/Login"},
		handler,
	)

	require.NoError(t, err)
	assert.True(t, handlerCalled)
	assert.Equal(t, "ok", resp)
}

func TestAuthInterceptor_ProtectedMethod_ValidToken(t *testing.T) {
	interceptor, jwtManager := newTestInterceptor()

	userID := uuid.New()
	token, err := jwtManager.GenerateToken(userID, "test@example.com")
	require.NoError(t, err)

	md := metadata.Pairs("authorization", "Bearer "+token)
	ctx := metadata.NewIncomingContext(context.Background(), md)

	var capturedCtx context.Context
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		capturedCtx = ctx
		return "ok", nil
	}

	unary := interceptor.Unary()
	resp, err := unary(
		ctx,
		nil,
		&grpc.UnaryServerInfo{FullMethod: "/vsevolodryzhov.gkeeper.proto.GKeeper/CreateSecret"},
		handler,
	)

	require.NoError(t, err)
	assert.Equal(t, "ok", resp)

	ctxUserID, ok := capturedCtx.Value(ctxKeyUserID).(uuid.UUID)
	require.True(t, ok)
	assert.Equal(t, userID, ctxUserID)

	ctxEmail, ok := capturedCtx.Value(ctxKeyEmail).(string)
	require.True(t, ok)
	assert.Equal(t, "test@example.com", ctxEmail)
}

func TestAuthInterceptor_ProtectedMethod_NoMetadata(t *testing.T) {
	interceptor, _ := newTestInterceptor()

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		t.Fatal("handler should not be called")
		return nil, nil
	}

	unary := interceptor.Unary()
	_, err := unary(
		context.Background(),
		nil,
		&grpc.UnaryServerInfo{FullMethod: "/vsevolodryzhov.gkeeper.proto.GKeeper/CreateSecret"},
		handler,
	)

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
}

func TestAuthInterceptor_ProtectedMethod_NoToken(t *testing.T) {
	interceptor, _ := newTestInterceptor()

	md := metadata.Pairs()
	ctx := metadata.NewIncomingContext(context.Background(), md)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		t.Fatal("handler should not be called")
		return nil, nil
	}

	unary := interceptor.Unary()
	_, err := unary(
		ctx,
		nil,
		&grpc.UnaryServerInfo{FullMethod: "/vsevolodryzhov.gkeeper.proto.GKeeper/GetSecrets"},
		handler,
	)

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
}

func TestAuthInterceptor_ProtectedMethod_InvalidToken(t *testing.T) {
	interceptor, _ := newTestInterceptor()

	md := metadata.Pairs("authorization", "Bearer invalid-token")
	ctx := metadata.NewIncomingContext(context.Background(), md)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		t.Fatal("handler should not be called")
		return nil, nil
	}

	unary := interceptor.Unary()
	_, err := unary(
		ctx,
		nil,
		&grpc.UnaryServerInfo{FullMethod: "/vsevolodryzhov.gkeeper.proto.GKeeper/GetSecrets"},
		handler,
	)

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
}

func TestAuthInterceptor_ProtectedMethod_BadFormat(t *testing.T) {
	interceptor, _ := newTestInterceptor()

	md := metadata.Pairs("authorization", "just-a-token-no-bearer")
	ctx := metadata.NewIncomingContext(context.Background(), md)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		t.Fatal("handler should not be called")
		return nil, nil
	}

	unary := interceptor.Unary()
	_, err := unary(
		ctx,
		nil,
		&grpc.UnaryServerInfo{FullMethod: "/vsevolodryzhov.gkeeper.proto.GKeeper/GetSecrets"},
		handler,
	)

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
	assert.Contains(t, st.Message(), "invalid authorization header format")
}

func TestIsProtectedMethod(t *testing.T) {
	interceptor, _ := newTestInterceptor()

	assert.False(t, interceptor.isProtectedMethod("/vsevolodryzhov.gkeeper.proto.GKeeper/Register"))
	assert.False(t, interceptor.isProtectedMethod("/vsevolodryzhov.gkeeper.proto.GKeeper/Login"))
	assert.True(t, interceptor.isProtectedMethod("/vsevolodryzhov.gkeeper.proto.GKeeper/CreateSecret"))
	assert.True(t, interceptor.isProtectedMethod("/vsevolodryzhov.gkeeper.proto.GKeeper/GetSecrets"))
	assert.True(t, interceptor.isProtectedMethod("/vsevolodryzhov.gkeeper.proto.GKeeper/DeleteSecret"))
}
