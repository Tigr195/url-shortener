package grpc_test

import (
	"context"
	"errors"
	"testing"

	urlpb "github.com/Tigr195/url-shortener/gen/url"
	grpchandler "github.com/Tigr195/url-shortener/internal/grpc"
	"github.com/Tigr195/url-shortener/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type mockURLService struct {
	mock.Mock
}

func (m *mockURLService) Shorten(ctx context.Context, url string) (*model.ShortenResponse, error) {
	args := m.Called(ctx, url)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ShortenResponse), args.Error(1)
}

func (m *mockURLService) Resolve(ctx context.Context, code string) (*model.URL, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.URL), args.Error(1)
}

func TestGRPCShorten_Success(t *testing.T) {
	svc := new(mockURLService)
	srv := grpchandler.NewURLGRPCServer(svc)

	svc.On("Shorten", mock.Anything, "https://google.com").
		Return(&model.ShortenResponse{
			ShortURL:    "http://localhost:8080/abc123",
			OriginalURL: "https://google.com",
		}, nil)

	resp, err := srv.Shorten(context.Background(), &urlpb.ShortenRequest{
		Url: "https://google.com",
	})

	assert.NoError(t, err)
	assert.Equal(t, "http://localhost:8080/abc123", resp.ShortUrl)
	assert.Equal(t, "https://google.com", resp.OriginalUrl)
}

func TestGRPCShorten_EmptyURL(t *testing.T) {
	svc := new(mockURLService)
	srv := grpchandler.NewURLGRPCServer(svc)

	_, err := srv.Shorten(context.Background(), &urlpb.ShortenRequest{
		Url: "",
	})

	assert.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

func TestGRPCShorten_ServiceError(t *testing.T) {
	svc := new(mockURLService)
	srv := grpchandler.NewURLGRPCServer(svc)

	svc.On("Shorten", mock.Anything, "https://google.com").
		Return(nil, errors.New("service error"))

	_, err := srv.Shorten(context.Background(), &urlpb.ShortenRequest{
		Url: "https://google.com",
	})

	assert.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
}

func TestGRPCResolve_Success(t *testing.T) {
	svc := new(mockURLService)
	srv := grpchandler.NewURLGRPCServer(svc)

	svc.On("Resolve", mock.Anything, "abc123").
		Return(&model.URL{
			ShortCode:   "abc123",
			OriginalURL: "https://google.com",
		}, nil)

	resp, err := srv.Resolve(context.Background(), &urlpb.ResolveRequest{
		Code: "abc123",
	})

	assert.NoError(t, err)
	assert.Equal(t, "https://google.com", resp.OriginalUrl)
}

func TestGRPCResolve_EmptyCode(t *testing.T) {
	svc := new(mockURLService)
	srv := grpchandler.NewURLGRPCServer(svc)

	_, err := srv.Resolve(context.Background(), &urlpb.ResolveRequest{
		Code: "",
	})

	assert.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

func TestGRPCResolve_NotFound(t *testing.T) {
	svc := new(mockURLService)
	srv := grpchandler.NewURLGRPCServer(svc)

	svc.On("Resolve", mock.Anything, "notexist").
		Return(nil, errors.New("not found"))

	_, err := srv.Resolve(context.Background(), &urlpb.ResolveRequest{
		Code: "notexist",
	})

	assert.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.NotFound, st.Code())
}
