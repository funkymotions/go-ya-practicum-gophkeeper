package apperror

import "google.golang.org/grpc/codes"

var StorageCreateBlockError = AppError{
	Message:    "failed to create storage block",
	GRPCStatus: codes.Internal,
}

var StorageListDataBlockError = AppError{
	Message:    "failed to list data blocks",
	GRPCStatus: codes.Internal,
}

var StorageErrorGeneric = AppError{
	Message:    "generic storage error",
	GRPCStatus: codes.Internal,
}

var StorageErrorNotFound = AppError{
	Message:    "data not found",
	GRPCStatus: codes.NotFound,
}

var StorageReadBlockTypeError = AppError{
	Message:    "failed to read block type",
	GRPCStatus: codes.Internal,
}
