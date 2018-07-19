package auth

// ArgumentError represents a user input error.
type ArgumentError struct {
	Source error
}

// Error gets string representation of ArgumentError.
func (e ArgumentError) Error() string {
	return e.Source.Error()
}

// ExpiredError indicates an object was expired and a new one was regenerated.
type ExpiredError struct {
	Message string
}

// Error gets string representation of ExpiredError.
func (e ExpiredError) Error() string {
	return e.Message
}

// NotFoundError indicates an object requested was not found.
type NotFoundError struct {
	Message string
}

// Error gets string representation of NotFoundError.
func (e NotFoundError) Error() string {
	return e.Message
}

// UnauthorizedError indicates authentication error.
type UnauthorizedError struct {
	Message string
}

// Error gets string representation of UnauthorizedError.
func (e UnauthorizedError) Error() string {
	return e.Message
}
