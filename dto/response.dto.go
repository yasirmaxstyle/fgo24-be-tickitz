package dto

type ErrorResponse struct {
	Success bool
	Error   error
}

type SuccessResponse struct {
	Success bool
	Message string
	Data    interface{}
}
