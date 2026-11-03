package services

type Error struct {
	Status  int
	Message string
}

// Error Return the domain error message
func (e *Error) Error() string {
	return e.Message
}

// fail Create a domain error with its corresponding HTTP status
func fail(status int, s string) error {
	return &Error{status, s}
}
