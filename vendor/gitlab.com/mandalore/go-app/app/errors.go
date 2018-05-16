package app

import (
	"encoding/json"
	"fmt"
)

// These are sample application errors.
const (
	// ErrorBadRequest should be used as when a request to a function is invalid and there is no other error code which makes more sense.
	ErrorBadRequest = 4000
	// ErrorInvalidParameter should be used when a function parameter has an invalid value.
	ErrorInvalidParameter = 4001
	// ErrorConflict should be used when attempting to save an entity which has a duplicate with conflicting information.
	ErrorConflict = 4090
	// ErrorNotFound should be used when any resource does not exist. It is a generic version of the entity not found.
	ErrorNotFound = 4040
	// ErrorDataInconsistency should be used when the data layer contains unexpected errors or inconsistencies.
	ErrorDataInconsistency = 5001
	// ErrorUnexpected should be used when the error is caused by third party libraries which means that the error cause should be set.
	ErrorUnexpected = 5002
	// ErrorDevPoo should be used when the most likely cause is developer error.
	ErrorDevPoo = 5003
	// ErrorConcurrencyException should be used when an operation fails due to any concurrency issues (deadlock, another operation already running and can only have one, etc).
	ErrorConcurrencyException = 5004
	// ErrorTimeout should be used when an operation exceeds the expected time.
	ErrorTimeout = 5005
	// ErrorOverflow should be used when a buffer, queue, stack, etc limit is exceeded.
	ErrorOverflow = 5006
	// ErrorBufferFull should be used when a buffer is full on a request to add data to the buffer.
	ErrorBufferFull = 5007
	// ErrorBufferEmpty should be used when a buffer is empty on a request to fetch data from a buffer
	ErrorBufferEmpty = 5008
)

// Error interface for wrapping application errors with error codes for faster error type switching.
type Error interface {
	error
	GetCode() int
	GetCause() error
}

// ErrorData interface is a an extension of the Error interface in order to provide structured data associated with an error. Useful for including identifiers and relevant information about the error.
type ErrorData interface {
	Error
	GetData() map[string]interface{}
}

// ApplicationError is a generic implementation of a serializable application
// Error.The Code should be used instead of the usual type reflection for
// handling different types of error. There is also a Cause which is not
// serialized but can be used to wrap original cause of errors. This in turn
// allows a logger to go through the cause stack and print the whole trace
// of errors. You should always wrap errors each time you handle them and
// log the error stack when the error is irrelevant for the upstream.
type ApplicationError struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Cause   error                  `json:"-"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

// NewError creates a new error struct which implements the Error interface.
func NewError(code int, message string, cause error) Error {
	return &ApplicationError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// NewErrorf is a wrapper to facilitate creating error messages with unstructured context data.
func NewErrorf(code int, cause error, format string, args ...interface{}) Error {
	return &ApplicationError{
		Cause:   cause,
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}

// NewErrorData creates a new error which implements the ErrorData interface.
func NewErrorData(code int, message string, cause error, data map[string]interface{}) ErrorData {
	return &ApplicationError{
		Cause:   cause,
		Message: message,
		Code:    code,
		Data:    data,
	}
}

// NewApplicationError creates a new ApplicationError.
// @deprecated
func NewApplicationError(code int, message string, cause error) *ApplicationError {
	return &ApplicationError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// NewUnexpectedError is a helper for generating an ApplicationError with error code set to ErrorUnexpected.
func NewUnexpectedError(message string, cause error) *ApplicationError {
	return &ApplicationError{
		Code:    ErrorUnexpected,
		Message: message,
		Cause:   cause,
	}
}

// Error ...
func (err *ApplicationError) Error() string {
	return err.Message
}

// GetCode ...
func (err *ApplicationError) GetCode() int {
	return err.Code
}

// GetCause ...
func (err *ApplicationError) GetCause() error {
	return err.Cause
}

// GetData ...
func (err *ApplicationError) GetData() map[string]interface{} {
	return err.Data
}

// StringifyError returns an error message, if the passed error is an
// app.Error then the full error stack is returned where each cause is
// separated by ';'.
func StringifyError(err error) string {
	if err == nil {
		return ""
	}

	if err, ok := err.(Error); ok {
		cause := err.GetCause()
		msg := err.Error()
		for cause != nil {
			msg += "; " + cause.Error()
			if tmp, ok := cause.(Error); ok {
				cause = tmp.GetCause()
			} else {
				cause = nil
			}
		}
		return msg
	}

	return err.Error()
}

// ErrorCauses returns a slice of all error causes. If error does not implement the Error interface then a nil slice is returned.
func ErrorCauses(err error) []string {
	if err, ok := err.(Error); ok {
		var causes = []string{}
		cause := err.GetCause()
		for cause != nil {
			causes = append(causes, cause.Error())
			if tmp, ok := cause.(Error); ok {
				cause = tmp.GetCause()
			} else {
				cause = nil
			}
		}
		return causes
	}

	return nil
}

// GetErrorStack returns a slice of the error stack where each error is converted to string.
func GetErrorStack(err error) []string {
	if err, ok := err.(Error); ok {
		var causes = []string{serializeError(err)}
		cause := err.GetCause()
		for cause != nil {
			if tmp, ok := cause.(Error); ok {
				causes = append(causes, serializeError(tmp))
				cause = tmp.GetCause()
			} else {
				causes = append(causes, serializeError(&ApplicationError{Code: -1, Message: cause.Error()}))
				cause = nil
			}
		}
		return causes
	}

	return []string{serializeError(&ApplicationError{Code: -1, Message: err.Error()})}
}

// GetErrorData checks the provided error and all subsequent error causes, in the case the error implements the Error interface, for the first error which implements the ErrorData interface and returns the associated data. If no data is found then nil is returned instead.
func GetErrorData(err error) map[string]interface{} {
	if err == nil {
		return nil
	}

	if err, ok := err.(ErrorData); ok {
		return err.GetData()
	}

	if err, ok := err.(Error); ok {
		cause := err.GetCause()
		for cause != nil {
			if tmp, ok := cause.(ErrorData); ok {
				return tmp.GetData()
			} else if tmp, ok := cause.(ErrorData); ok {
				cause = tmp.GetCause()
			} else {
				cause = nil
			}
		}
		return nil
	}

	return nil
}

// MergeWithErrorData merges the provided data with the error's data, if any. This will overwrite any map key with the same name.
func MergeWithErrorData(data map[string]interface{}, err error) map[string]interface{} {
	if err == nil {
		return data
	}

	eData := GetErrorData(err)

	if eData != nil {
		if data == nil {
			return eData
		}
		for k, v := range eData {
			data[k] = v
		}
	}

	return data
}

func serializeError(err Error) string {
	str, e := json.Marshal(err)
	if e != nil {
		return err.Error()
	}

	return string(str)
}
