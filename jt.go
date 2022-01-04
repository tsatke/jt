package jt

type Error string

func (e Error) Error() string {
	return string(e)
}

const (
	ErrNotImplemented     Error = "not implemented"
	ErrUnknownProjectKind Error = "project kind is unknown"
)
