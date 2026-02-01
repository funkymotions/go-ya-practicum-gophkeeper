package apperror

import (
	"fmt"

	"google.golang.org/grpc/codes"
)

func (e *AppError) Error() string {
	return fmt.Sprintf("code: %d, %s", e.GRPCStatus, e.Message)
}

var AuthInvalidCredentialsError = &AppError{
	Message:    "invalid credentials",
	GRPCStatus: codes.Unauthenticated,
}

var AuthUserNotExistsError = &AppError{
	Message:    "user does not exist",
	GRPCStatus: codes.NotFound,
}

var AuthUserExistsError = &AppError{
	Message:    "user already exists",
	GRPCStatus: codes.AlreadyExists,
}

var AuthCreateUserError = &AppError{
	Message:    "failed to create user",
	GRPCStatus: codes.Internal,
}

var AuthErrorGeneric = &AppError{
	Message:    "generic authentication error",
	GRPCStatus: codes.Internal,
}
