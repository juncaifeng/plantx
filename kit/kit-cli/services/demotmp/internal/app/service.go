package app

import "context"

type Service struct{}

func New() *Service {
	return &Service{}
}

func (s *Service) Ping(ctx context.Context, msg string) string {
	return msg
}
