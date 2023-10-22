package db

import "errors"

var ( 
	ErrConflict = errors.New("data conflict")
	ErrNoRows = errors.New("no rows")
	ErrRowExist = errors.New("row already exists")
	ErrUserNotExist = errors.New("user does not exist")
)