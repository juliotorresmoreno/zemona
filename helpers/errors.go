package helpers

type httpError struct {
	Message string
}

func MakeHTTPError(err error) *httpError {
	return &httpError{
		Message: err.Error(),
	}
}
