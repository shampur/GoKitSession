package session

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/sessions"
	"golang.org/x/net/context"
)

//Service Interface of session manager
type Service interface {
	login(ctx context.Context, req LoginRequest) (LoginResponse, error)
}

// LoginRequest service
type LoginRequest struct {
	httpreq *http.Request
	cred    Credentials
}

//
// LoginResponse service
type LoginResponse struct {
	Authenticated bool              `json:"authenticated"`
	Message       string            `json:"message"`
	Session       *sessions.Session `json:"session"`
	Httpreq       *http.Request     `json:"httpreq"`
}

//Credentials object of the user
type Credentials struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	Organization string `json:"organization"`
}

type sessionService struct {
	mtx         sync.RWMutex
	store       *sessions.CookieStore
	authManager *AuthManager
}

//NewSessionService contains the session store
func NewSessionService() Service {
	return &sessionService{
		store:       sessions.NewCookieStore([]byte("something-very-secret")),
		authManager: NewAuthmanager(),
	}
}

var (
	// ErrInconsistentIDs server error message
	ErrInconsistentIDs = errors.New("inconsistent IDs")
	// ErrAlreadyExists server error message
	ErrAlreadyExists = errors.New("already exists")
	// ErrNotFound server error message
	ErrNotFound = errors.New("not found")
)

func (s *sessionService) login(ctx context.Context, r LoginRequest) (LoginResponse, error) {
	fmt.Println("Login service called")
	s.mtx.Lock()
	defer s.mtx.Unlock()
	var res LoginResponse
	session, err := s.store.Get(r.httpreq, "contiv-session")

	if err != nil {
		fmt.Println("error while retrieving session info")
		return LoginResponse{}, err
	}

	if !session.IsNew {
		res, err = validate(session)
	} else {
		res, err = s.authManager.authenticate(r.cred)
	}
	if res.Authenticated {
		session.Values["Username"] = r.cred.Username
		session.Values["LastLoginTime"] = time.Now().Format(time.RFC3339)
		res.Session = session
		res.Httpreq = r.httpreq
	}
	return res, err
}

func validate(session *sessions.Session) (LoginResponse, error) {
	fmt.Println("validate session called")
	fmt.Println("The session values are=")
	for k, v := range session.Values {
		fmt.Println("key=", k, "values", v)
	}
	lastLoginTime := session.Values["LastLoginTime"].(string)
	fmt.Println("LastLoginTime = ", lastLoginTime)
	parsedTime, err := time.Parse(time.RFC3339, lastLoginTime)
	if err != nil {
		fmt.Println("Error while parsing time")
		return LoginResponse{}, err
	}
	duration := time.Since(parsedTime)
	fmt.Println("currentTime = ", time.Now().Format(time.RFC3339))
	fmt.Println("Past time = ", parsedTime)
	fmt.Println("Duration eloped = ", duration.Minutes())
	minutesPassed := duration.Minutes()
	if minutesPassed < 0 || minutesPassed > 0.3 {
		fmt.Println("Session Invalid")
		return LoginResponse{Authenticated: false, Message: "Invalid Session"}, nil
	}
	session.Values["LastLoginTime"] = time.Now().Format(time.RFC3339)
	fmt.Println("Session Valid")
	return LoginResponse{Authenticated: true, Message: "Success"}, nil
}

// Au
