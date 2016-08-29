package session

import "fmt"

//AuthManager is responsible for cycling through the different authentication mechanizms
type AuthManager struct {
	authModuleCount int64
	authModules     []AuthInterface
}

func (a *AuthManager) authenticate(cred Credentials) (LoginResponse, error) {
	fmt.Println("Inside auth manager authenticate")
	var result LoginResponse
	var err error
	if len(cred.Organization) > 0 && len(cred.Username) > 0 && len(cred.Password) > 0 {
		result, err = a.authModules[0].authenticate(cred)
		return result, err
	}
	return LoginResponse{Authenticated: false, Message: "Credential or username not correct"}, nil
}

//NewAuthmanager creates a new authentication manager
func NewAuthmanager() *AuthManager {
	return &AuthManager{
		authModuleCount: 1,
		authModules:     createModules(),
	}
}

func createModules() []AuthInterface {
	var inter []AuthInterface
	ldapModule := newLdap()
	inter = append(inter, ldapModule)
	return inter
}

//AuthInterface - all authentication modules should implement this interface
type AuthInterface interface {
	authenticate(cred Credentials) (LoginResponse, error)
}

type ldap struct {
	ldapType    string
	ldapVersion string
}

func newLdap() *ldap {
	return &ldap{
		ldapType:    "ldap-server1",
		ldapVersion: "ldap3v",
	}
}

func (l *ldap) authenticate(cred Credentials) (LoginResponse, error) {
	fmt.Println("Inside ldap authenticate")
	fmt.Println("username=", cred.Username, "Password=", cred.Password)
	if cred.Username == "contiv" && cred.Password == "123" {
		return LoginResponse{Authenticated: true, Message: "success"}, nil
	}
	return LoginResponse{Authenticated: false, Message: "Invalid username or password"}, nil
}
