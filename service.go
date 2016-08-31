package session

import (
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
	logout(ctx context.Context, req LogoutRequest) (LogoutResponse, error)
	validateapp(ctx context.Context, req validateAppRequest) (LoginResponse, error)
}

//validate app request
type validateAppRequest struct {
	httpreq *http.Request
}

//LogoutRequest
type LogoutRequest struct {
	httpreq *http.Request
}

// LoginRequest service
type LoginRequest struct {
	httpreq *http.Request
	cred    Credentials
}

// LogoutResponse service
type LogoutResponse struct {
	Session       *sessions.Session `json:"session"`
	Httpreq       *http.Request     `json:"httpreq"`
}
// LoginResponse service
type LoginResponse struct {
	Authenticated bool              `json:"authenticated"`
	Message       string            `json:"message"`
	Username      string		`json:"username"`
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
	store       *sessions.FilesystemStore
	authManager *AuthManager
}

//NewSessionService contains the session store
func NewSessionService() Service {
	return &sessionService{
		store:       sessions.NewFilesystemStore("./sessionstore",[]byte("something-very-secret")),
		authManager: NewAuthmanager(),
	}
}

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
		if err != nil {
			return LoginResponse{}, err
		}
		if res.Authenticated != true {
			res, err = s.authManager.authenticate(r.cred)
		}
	} else {
		res, err = s.authManager.authenticate(r.cred)
	}
	if res.Authenticated {
		session.Values["Username"] = r.cred.Username
		session.Values["LastLoginTime"] = time.Now().Format(time.RFC3339)
		res.Username = r.cred.Username
	} else {
		session.Options.MaxAge = -1
	}
	res.Session = session
	res.Httpreq = r.httpreq
	return res, err
}

func (s *sessionService) logout(ctx context.Context, r LogoutRequest) (LogoutResponse, error) {
	fmt.Println("Logout service called")
	s.mtx.Lock()
	defer s.mtx.Unlock()
	var res LogoutResponse
	session, err := s.store.Get(r.httpreq, "contiv-session")

	if err != nil {
		fmt.Println("error while retrieving session info")
		return LogoutResponse{}, err
	}

	session.Options.MaxAge = -1
	res.Session = session
	res.Httpreq =r.httpreq

	return res, err
}

func (s *sessionService) validateapp(ctx context.Context, r validateAppRequest) (LoginResponse, error) {
	fmt.Println("validate app request service called")
	s.mtx.Lock()
	defer s.mtx.Unlock()
	var res LoginResponse
	session, err := s.store.Get(r.httpreq, "contiv-session")
	if err != nil {
		fmt.Println("error while retrieving session info")
		return LoginResponse{}, err
	}
	if (session.IsNew) {
		fmt.Println("session is new")
		res.Authenticated = false
		res.Message = "Invalid Session validateapp"
		session.Options.MaxAge = -1
	} else {
		fmt.Println("session is present")
		res, err = validate(session)
		if res.Authenticated {
			res.Username = session.Values["Username"].(string)
		}
	}

	res.Session = session
	res.Httpreq = r.httpreq

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
	if minutesPassed < 0 || minutesPassed > 0.20 {
		fmt.Println("Session Invalid")
		session.Options.MaxAge = -1
		return LoginResponse{Authenticated: false, Message: "Invalid Session"}, nil
	}
	session.Values["LastLoginTime"] = time.Now().Format(time.RFC3339)
	fmt.Println("Session Valid")
	return LoginResponse{Authenticated: true, Message: "Success"}, nil
}

// Au
