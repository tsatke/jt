package jt

import "github.com/rs/zerolog"

var (
	log = zerolog.Nop()
)

func SetLogger(newLogger zerolog.Logger) {
	log = newLogger
}

type Error string

func (e Error) Error() string {
	return string(e)
}

const (
	ErrNotImplemented     Error = "not implemented"
	ErrUnknownProjectKind Error = "project kind is unknown"
)
