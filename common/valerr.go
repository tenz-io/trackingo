package common

import "errors"

type ValError struct {
	Code int
	Err  error
}

func NewValError(code int, err error) *ValError {
	return &ValError{
		Code: code,
		Err:  err,
	}
}

func (ve *ValError) Error() string {
	if ve.Err == nil {
		return ""
	}
	return ve.Err.Error()
}

// ErrorCode returns the error code of the given error.
// If the given error is nil, it returns 0.
// If the given error is not a ValError, it returns 1.
func ErrorCode(err error) int {
	if err == nil {
		return 0
	}

	var valErr *ValError
	if match := errors.As(err, &valErr); match {
		return valErr.Code
	}

	return 1
}

// ErrorMsg returns the error message of the given error.
func ErrorMsg(err error) string {
	if err == nil {
		return ""
	}

	return err.Error()
}
