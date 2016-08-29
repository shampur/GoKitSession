package session

import (
	"golang.org/x/net/context"

	"github.com/go-kit/kit/endpoint"
)

// Endpoints exposed by the service
type Endpoints struct {
	loginEndpoint endpoint.Endpoint
}

// MakeServerEndpoints function prepares the server Endpoints
func MakeServerEndpoints(s Service) Endpoints {
	return Endpoints{
		loginEndpoint: MakeLoginEnpoint(s),
	}
}

// MakeLoginEnpoint prepars the login endpoint
func MakeLoginEnpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(LoginRequest)
		result, err := s.login(ctx, req)
		return result, err
	}
}
