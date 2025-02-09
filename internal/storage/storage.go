package storage

import "errors"

var(
	ErrURLNotFound = errors.New("URL not found")
	ErrURLExists = errors.New("URL already exists in the database")
	ErrAliasExists = errors.New("alias already exists in the database")
)
