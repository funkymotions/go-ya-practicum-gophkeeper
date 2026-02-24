package apperror

import "google.golang.org/grpc/codes"

var TypesReadError = &AppError{
	Message:    "failed to read block type",
	GRPCStatus: codes.Internal,
}
