package grpc

import (
	"context"
	"fmt"

	"github.com/Tigr195/url-shortener/gen/url"
	"github.com/Tigr195/url-shortener/internal/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type URLService interface {
	Shorten(ctx context.Context, originalURL string) (*model.ShortenResponse, error)
	Resolve(ctx context.Context, code string) (*model.URL, error)
}

type URLGRPCServer struct {
	url.UnimplementedURLShortenerServer
	service URLService
}

func NewURLGRPCServer(service URLService) *URLGRPCServer {
	return &URLGRPCServer{service: service}
}

func (s *URLGRPCServer) Shorten(ctx context.Context, req *url.ShortenRequest) (*url.ShortenResponse, error) {
	if req.Url == "" {
		return nil, status.Error(codes.InvalidArgument, "url is required")
	}

	resp, err := s.service.Shorten(ctx, req.Url)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to shorten url: %v", err))
	}

	return &url.ShortenResponse{
		ShortUrl:    resp.ShortURL,
		OriginalUrl: resp.OriginalURL,
	}, nil
}

func (s *URLGRPCServer) Resolve(ctx context.Context, req *url.ResolveRequest) (*url.ResolveResponse, error) {
	if req.Code == "" {
		return nil, status.Error(codes.InvalidArgument, "code is required")
	}

	u, err := s.service.Resolve(ctx, req.Code)
	if err != nil {
		return nil, status.Error(codes.NotFound, "url not found")
	}

	return &url.ResolveResponse{
		OriginalUrl: u.OriginalURL,
	}, nil
}
