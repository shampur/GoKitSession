package session

import "fmt"

type ldapAuth struct {
	ldapType    string
	ldapVersion string
}

// NewLdap function initializes the ldap module
func NewLdapAuth() *ldapAuth {
	return &ldapAuth{
		ldapType:    "ldap-server1",
		ldapVersion: "ldap3v",
	}
}

func (l *ldapAuth) authenticate(cred Credentials) (LoginResponse, error) {
	fmt.Println("Inside ldap authenticate")
	fmt.Println("username=", cred.Username, "Password=", cred.Password)
	if cred.Username == "contiv" && cred.Password == "123" {
		return LoginResponse{Authenticated: true, Message: "success"}, nil
	}
	return LoginResponse{Authenticated: false, Message: "Invalid username or password"}, nil
}
