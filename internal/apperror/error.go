package apperror

import "google.golang.org/grpc/codes"

type AppError struct {
	Message    string
	GRPCStatus codes.Code
}
