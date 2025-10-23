package service

import (
	"context"
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
)

type Service struct {
	srv micro.Service
}

type Config struct {
	Name      string
	Version   string
	Endpoints map[string]func() micro.Handler
}

func NewService(ctx context.Context, nc *nats.Conn, cfg Config) (*Service, error) {
	srv, err := micro.AddService(nc, micro.Config{
		Name:    cfg.Name,
		Version: cfg.Version,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create micro service: %w", err)
	}

	for endpoint, handlerFunc := range cfg.Endpoints {
		h := handlerFunc()
		if err := srv.AddEndpoint(endpoint, h); err != nil {
			return nil, fmt.Errorf("failed to add handler for endpoint '%s': %w", endpoint, err)
		}
	}

	return &Service{
		srv: srv,
	}, nil
}

func (s *Service) Stop() error {
	return s.srv.Stop()
}

func Call(ctx context.Context, nc *nats.Conn, serviceName, endpoint string, req []byte) ([]byte, error) {
	msg := nats.NewMsg(fmt.Sprintf("%s.%s", serviceName, endpoint))
	msg.Data = req

	resp, err := nc.RequestMsgWithContext(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to call service '%s' endpoint '%s': %w", serviceName, endpoint, err)
	}

	return resp.Data, nil
}
