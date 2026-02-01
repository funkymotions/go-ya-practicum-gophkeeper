package apperror

type DBError struct {
	Message string
}

func (e *DBError) Error() string {
	return e.Message
}

var DBErrorNoRows = &DBError{Message: "no rows in result set"}
