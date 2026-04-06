package service_errors

import (
	"fmt"
)

type ServiceError struct {
	Code    int
	Err     error
	Message string
}

func (e *ServiceError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}

	return e.Message
}

func (e *ServiceError) Unwrap() error {
	return e.Err
}

type NotFoundError struct {
	ServiceError
}

func NewNotFoundError(message string) *NotFoundError {
	return &NotFoundError{
		ServiceError: ServiceError{
			Code:    404,
			Message: message,
		},
	}
}

type ConflictError struct {
	ServiceError
}

func NewConflictError(message string) *ConflictError {
	return &ConflictError{
		ServiceError: ServiceError{
			Code:    409,
			Message: message,
		},
	}
}

type BadRequestError struct {
	ServiceError
}

func NewBadRequestError(message string) *BadRequestError {
	return &BadRequestError{
		ServiceError: ServiceError{
			Code:    400,
			Message: message,
		},
	}
}

type UnauthorizedError struct {
	ServiceError
}

func NewUnauthorizedError(message string) *UnauthorizedError {
	return &UnauthorizedError{
		ServiceError: ServiceError{
			Code:    401,
			Message: message,
		},
	}
}

type ForbiddenError struct {
	ServiceError
}

func NewForbiddenError(message string) *ForbiddenError {
	return &ForbiddenError{
		ServiceError: ServiceError{
			Code:    403,
			Message: message,
		},
	}
}

type InternalServerError struct {
	ServiceError
}

func NewInternalServerError(message string) *InternalServerError {
	return &InternalServerError{
		ServiceError: ServiceError{
			Code:    500,
			Message: message,
		},
	}
}
