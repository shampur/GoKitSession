package session

import (
	"fmt"
	"time"

	"golang.org/x/net/context"

	"github.com/go-kit/kit/log"
)

// Middleware describes a service (as opposed to endpoint) middleware.
type Middleware func(Service) Service

//LoggingMiddleware for logging services execution
func LoggingMiddleware(logger log.Logger) Middleware {
	fmt.Println("Logging middleware")
	return func(next Service) Service {
		return &loggingMiddleware{
			next:   next,
			logger: logger,
		}
	}
}

type loggingMiddleware struct {
	next   Service
	logger log.Logger
}

func (mw loggingMiddleware) login(ctx context.Context, r LoginRequest) (resp LoginResponse, err error) {
	defer func(begin time.Time) {
		mw.logger.Log("Host", r.httpreq.Host, "Url", r.httpreq.URL)
		mw.logger.Log("method", "login", "took", time.Since(begin), "err", err)
	}(time.Now())
	fmt.Println("login middleware called")
	resp, err = mw.next.login(ctx, r)
	return
}

func (mw loggingMiddleware) logout(ctx context.Context, r LogoutRequest) (resp LogoutResponse, err error) {
	defer func(begin time.Time) {
		mw.logger.Log("Host", r.httpreq.Host, "Url", r.httpreq.URL)
		mw.logger.Log("method", "Logout", "took", time.Since(begin), "err", err)
	}(time.Now())
	fmt.Println("logout middleware called")
	resp, err = mw.next.logout(ctx, r)
	return
}

func (mw loggingMiddleware) validateapp(ctx context.Context, r validateAppRequest) (resp LoginResponse, err error) {
	defer func(begin time.Time) {
		mw.logger.Log("Host", r.httpreq.Host, "Url", r.httpreq.URL)
		mw.logger.Log("method", "validateapp", "took", time.Since(begin), "err", err)
	}(time.Now())
	fmt.Println("validateapp middleware called")
	resp, err = mw.next.validateapp(ctx, r)
	return
}
