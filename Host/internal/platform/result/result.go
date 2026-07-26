package result

import "net/http"

type FieldErrors map[string][]string

type AppError struct {
	Code        string
	Message     string
	Status      int
	FieldErrors FieldErrors
	Cause       error
}

func (e AppError) Error() string {
	return e.Message
}

func Invalid(code, message string, fields FieldErrors) AppError {
	return AppError{Code: code, Message: message, Status: http.StatusBadRequest, FieldErrors: fields}
}

func Forbidden(code, message string) AppError {
	return AppError{Code: code, Message: message, Status: http.StatusForbidden}
}

func Conflict(code, message string) AppError {
	return AppError{Code: code, Message: message, Status: http.StatusConflict}
}

func Internal(cause error) AppError {
	return AppError{
		Code:    "internal.unexpected",
		Message: "Something went wrong. Please try again.",
		Status:  http.StatusInternalServerError,
		Cause:   cause,
	}
}

type Outcome[T any] struct {
	Value T
	Err   *AppError
}

func Ok[T any](value T) Outcome[T] {
	return Outcome[T]{Value: value}
}

func Fail[T any](err AppError) Outcome[T] {
	return Outcome[T]{Err: &err}
}

func Map[A, B any](source Outcome[A], fn func(A) B) Outcome[B] {
	if source.Err != nil {
		return Outcome[B]{Err: source.Err}
	}
	return Ok(fn(source.Value))
}

func Bind[A, B any](source Outcome[A], fn func(A) Outcome[B]) Outcome[B] {
	if source.Err != nil {
		return Outcome[B]{Err: source.Err}
	}
	return fn(source.Value)
}
