package repo

import "errors"

var ErrNotFound = errors.New("not found")

var ErrExists = errors.New("already exists")
