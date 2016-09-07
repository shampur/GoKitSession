package session

import (
	"fmt"
	"net/http"
	"sync"
	"time"
	"errors"
	"strings"

	"github.com/gorilla/sessions"
	"golang.org/x/net/context"

	"io/ioutil"
)

//Service Interface of session manager
type Service interface {
	login(ctx context.Context, req LoginRequest) (LoginResponse, error)
	logout(ctx context.Context, req LogoutRequest) (LogoutResponse, error)
	validateapp(ctx context.Context, req validateAppRequest) (LoginResponse, error)
	apiprocess(ctx context.Context, req apiRequest)  (interface{}, error)
}

//validate app request
type validateAppRequest struct {
	httpreq *http.Request
}


type apiRequest struct {
	httpreq *http.Request
	data interface{}
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
	mtx         	sync.RWMutex
	store       	*sessions.FilesystemStore
	authmanager 	*AuthManager
	apiconfig	*apiConfig
}


//NewSessionService contains the session store
func NewSessionService() Service {
	return &sessionService{
		store:       	sessions.NewFilesystemStore("./sessionstore",[]byte("something-very-secret")),
		authmanager: 	NewAuthmanager(),
		apiconfig: 	GetApiConfig(),
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
			res, err = s.authmanager.authenticate(r.cred)
		}
	} else {
		res, err = s.authmanager.authenticate(r.cred)
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
			session.Values["LastLoginTime"] = time.Now().Format(time.RFC3339)
		}
	}

	res.Session = session
	res.Httpreq = r.httpreq

	return res, err
}

func (s *sessionService) apiprocess(ctx context.Context, r apiRequest) (interface{}, error) {
	fmt.Println("apiget request handler")
	s.mtx.Lock()
	defer s.mtx.Unlock()
	var result interface{}
	//var sesssionresp LoginResponse
	session, err := s.store.Get(r.httpreq, "contiv-session")
	if err != nil {
		fmt.Println("error while retrieving session")
		return result, err
	}

	if session.IsNew {
		fmt.Println("api process session is valid")
		result, err = apiexecute(s.apiconfig, r)
	}
	/*
	if session.IsNew {
		fmt.Println("api-get session is new")
		return result, nil
	} else {
		sesssionresp, err = validate(session)
		if sesssionresp.Authenticated {
			fmt.Println("api process session is valid")
			result, err = apiexecute(s.apiconfig, r)
		}

	}
	*/
	return result, err
}

func apiexecute(apiconfig *apiConfig, r apiRequest) (interface{}, error) {

	var result interface{}
	var err error

	config, ok := validateapi(apiconfig, r)

	if ok {
		switch r.httpreq.Method {

		case "GET": 	fmt.Println("The remote call =", config.Destination + r.httpreq.URL.Path)
				result, err = httpGet(config.Destination + r.httpreq.URL.Path)
				/*
				dump, err := httputil.DumpResponse(result, true)
				if err != nil {
					fmt.Println("error in dumping response")
				}
				fmt.Println(dump)
				*/
				return result, err
		case "POST":
		case "PUT":
		case "DELETE":

		}
	}
	return result, ErrNotFound
}

func validateapi(apiconfig *apiConfig, r apiRequest) (routedetail, bool) {
	for _, element := range (apiconfig.routelist) {
		if(strings.Contains(r.httpreq.URL.Path, element.Api)){
			if(contains(element.Methods, r.httpreq.Method) >= 0){
				return element, true
			}
		}
	}
	return routedetail{}, false
}

func contains(list interface{}, item interface{}) int {
	for index, element := range list.([]string) {
		if element == item {
			return index
		}
	}
	return -1
}

func httpGet(url string) (interface{}, error){

	r, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer r.Body.Close()

	switch {
	case r.StatusCode == int(404):
		return nil, errors.New("Page not found!")
	case r.StatusCode == int(403):
		return nil, errors.New("Access denied!")
	case r.StatusCode == int(500):
		response, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}
		return nil, errors.New(string(response))
	case r.StatusCode != int(200):
		return nil, errors.New(r.Status)
	}

	response, err := ioutil.ReadAll(r.Body)

	if err != nil {
		return nil, err
	}

	return response, nil
	//return r, nil
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
	fmt.Println("Session Valid")
	return LoginResponse{Authenticated: true, Message: "Success"}, nil
}

// Au
