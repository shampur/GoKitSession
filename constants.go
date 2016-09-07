package session

import "errors"

var (
	// ErrInconsistentIDs server error message
	ErrInconsistentIDs = errors.New("inconsistent IDs")
	// ErrAlreadyExists server error message
	ErrAlreadyExists = errors.New("already exists")
	// ErrNotFound server error message
	ErrNotFound = errors.New("not found")
)

var (
	//Session file location
	SessionDirectory = "sessionstore"
	LocalAuthFileLoc = "localauthfile.json"
	apiconfigfile = "apiconfig.json"
)
