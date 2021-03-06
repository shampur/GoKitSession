package session

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"golang.org/x/net/context"

	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/sessions"
	"os"
	"io"
)


//MakeHTTPHandler go http server
func MakeHTTPHandler(ctx context.Context, s Service, logger log.Logger) http.Handler {
	fmt.Println("Make http handler")
	r := mux.NewRouter()
	e := MakeServerEndpoints(s)
	options := []httptransport.ServerOption{
		httptransport.ServerErrorLogger(logger),
		httptransport.ServerErrorEncoder(encodeError),
	}

	// POST    /profiles/                          adds another profile
	// GET     /profiles/:id                       retrieves the given profile by id
	// PUT     /profiles/:id                       post updated profile information about the profile
	// PATCH   /profiles/:id                       partial updated profile information
	// DELETE  /profiles/:id                       remove the given profile
	// GET     /profiles/:id/addresses/            retrieve addresses associated with the profile
	// GET     /profiles/:id/addresses/:addressID  retrieve a particular profile address
	// POST    /profiles/:id/addresses/            add a new address
	// DELETE  /profiles/:id/addresses/:addressID  remove an address

	r.Methods("POST").Path("/loginvalidate/").Handler(httptransport.NewServer(
		ctx,
		e.loginEndpoint,
		decodeLoginReq,
		encodeLoginResponse,
		options...,
	))
	r.Methods("DELETE").Path("/logoutuser/").Handler(httptransport.NewServer(
		ctx,
		e.logoutEndpoint,
		decodeLogoutReq,
		encodeLogoutResponse,
		options...,
	))
	r.Methods("GET").Path("/validateapp/").Handler(httptransport.NewServer(
		ctx,
		e.validateappEndpoint,
		decodeValidateReq,
		encodeLoginResponse,
		options...,
	))
	r.PathPrefix("/").Handler(httptransport.NewServer(
		ctx,
		e.apiEndpoint,
		decodeApiRequest,
		encodeApiResponse,
		options...,
	))
	return r
}

func decodeLoginReq(_ context.Context, r *http.Request) (request interface{}, err error) {
	var req LoginRequest
	if e := json.NewDecoder(r.Body).Decode(&req.cred); e != nil {
		return nil, e
	}
	req.httpreq = r
	return req, nil
}

func decodeLogoutReq(_ context.Context, r *http.Request) (request interface{}, err error) {
	var req LogoutRequest
	req.httpreq = r
	return req, nil
}

func decodeValidateReq(_ context.Context, r *http.Request) (request interface{}, err error) {
	var req validateAppRequest
	req.httpreq = r
	return req, nil
}

func decodeApiRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	var req apiRequest
	req.httpreq = r

	if (r.Method == "POST" || r.Method == "PUT") {
		fmt.Println("r.body = ", r.Body)

		if e:= json.NewDecoder(r.Body).Decode(&req.data); e != nil {
			if(e != io.EOF) {
				return nil, e
			}

		}
	}

	return req, nil
}

// errorer is implemented by all concrete response types that may contain
// errors. It allows us to change the HTTP response code without needing to
// trigger an endpoint (transport-level) error. For more information, read the
// big comment in endpoints.go.
type errorer interface {
	error() error
}

// encodeResponse is the common method to encode all response types to the
// client. I chose to do it this way because, since we're using JSON, there's no
// reason to provide anything more specific. It's certainly possible to
// specialize on a per-response (per-method) basis.
func encodeLoginResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	fmt.Println("Login response function")
	type resp struct {
		Authenticated 	bool   	`json:"authenticated"`
		Message       	string 	`json:"message"`
		Username	string	`json:"username"`
	}
	if e, ok := response.(errorer); ok && e.error() != nil {
		// Not a Go kit transport error, but a business-logic error.
		// Provide those as HTTP errors.
		encodeError(ctx, e.error(), w)
		return nil
	}

	response.(LoginResponse).Session.Save(response.(LoginResponse).Httpreq, w)
	if(!response.(LoginResponse).Authenticated){
		delete(response.(LoginResponse).Session)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(resp{Authenticated: response.(LoginResponse).Authenticated,
		Message: response.(LoginResponse).Message, Username: response.(LoginResponse).Username})
	return nil
}

func encodeLogoutResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	fmt.Println("Logout Resposne function")
	if e, ok := response.(errorer); ok && e.error() != nil {
		encodeError(ctx, e.error(), w)
		return nil
	}
	response.(LogoutResponse).Session.Save(response.(LogoutResponse).Httpreq, w)
	delete(response.(LogoutResponse).Session)
	return nil
}

func encodeApiResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	fmt.Println("api encode resopnse called")
	fmt.Println("response = ", response)
	var jdata interface{}
	if e, ok := response.(errorer); ok && e.error() != nil {
		encodeError(ctx, e.error(), w)
		return nil
	}
	if((response != nil) && (len(response.([]byte)) > 0)){
		err := json.Unmarshal(response.([]byte), &jdata)
		if err != nil {
			fmt.Println("error in encoding response")
			encodeError(ctx, err, w)
			return nil
		}
		json.NewEncoder(w).Encode(jdata)
	}
	return nil
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	if err == nil {
		panic("encodeError with nil error")
	}
	//w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(codeFrom(err))
	io.WriteString(w, err.Error())
	/*
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
	*/
}

func delete(session *sessions.Session) {
	var filePath string
	filePath = SessionDirectory + "/" + "session_" + session.ID
	fmt.Println("filePath = ", filePath)
	// detect if file exists

	// delete file if it exists
	if _, err := os.Stat(filePath); err == nil {
		var err = os.Remove(filePath)
		if err != nil {
			return
		}
	}
}

func codeFrom(err error) int {
	switch err {
	case ErrNotFound:
		return http.StatusNotFound
	case ErrAlreadyExists, ErrInconsistentIDs:
		return http.StatusBadRequest
	default:
		if e, ok := err.(httptransport.Error); ok {
			switch e.Domain {
			case httptransport.DomainDecode:
				return http.StatusBadRequest
			case httptransport.DomainDo:
				return http.StatusServiceUnavailable
			default:
				return http.StatusInternalServerError
			}
		}
		return http.StatusInternalServerError
	}
}
