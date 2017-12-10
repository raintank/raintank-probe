package models

import (
	"fmt"
)

type AppError interface {
	Code() int
	Message() string
}

type ValidationError struct {
	Msg string
}

func NewValidationError(msg string) ValidationError {
	return ValidationError{Msg: msg}
}

func (e ValidationError) Code() int {
	return 400
}
func (e ValidationError) Message() string {
	return e.Msg
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%d: %s", 400, e.Msg)
}

type NotFoundError struct {
	Msg string
}

func NewNotFoundError(msg string) NotFoundError {
	return NotFoundError{Msg: msg}
}

func (e NotFoundError) Code() int {
	return 404
}
func (e NotFoundError) Message() string {
	return e.Msg
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("%d: %s", 404, e.Msg)
}
