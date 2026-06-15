package types

import (
	"errors"
	"fmt"
)

type ctxKey int

const (
	RequestIDKey ctxKey = iota
	UserIDKey
)

type AppErrorInterface interface {
	Error() string
	Unwrap() error
	PublicMessage() string
	Code() int
	Is(target error) bool
}

type AppError struct {
	Err        error
	PublicMsg  string
	StatusCode int
	Layer      string // "repo", "usecase", "handler"
	Op         string
}

func (e *AppError) Error() string {
	if e.Op != "" {
		return fmt.Sprintf("[%s:%s] %v", e.Layer, e.Op, e.Err)
	}
	return fmt.Sprintf("[%s] %v", e.Layer, e.Err)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func (e *AppError) PublicMessage() string {
	return e.PublicMsg
}

func (e *AppError) Code() int {
	return e.StatusCode
}

func (e *AppError) Is(target error) bool {
	return errors.Is(e.Err, target)
}
