package db

import "errors"

var ErrConflict = errors.New("data conflict")
var ErrNoRows = errors.New("no rows")
