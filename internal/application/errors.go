package application

type Code string

const (
	InvalidArgument    Code = "invalid_argument"
	NotFound           Code = "not_found"
	AlreadyExists      Code = "already_exists"
	RevisionConflict   Code = "revision_conflict"
	RequestConflict    Code = "request_conflict"
	ResourceExhausted  Code = "resource_exhausted"
	StorageFailure     Code = "storage_failure"
	CalculationFailure Code = "calculation_failure"
)

// Error описывает ошибку команды без привязки к транспортному протоколу.
type Error struct {
	Code    Code
	Message string
}

func (e *Error) Error() string { return string(e.Code) + ": " + e.Message }

func failure(code Code, message string) error { return &Error{Code: code, Message: message} }
