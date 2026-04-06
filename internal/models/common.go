package models

type DecodeRequestBodyError struct {
	Code        string
	Message     string
	FieldErrors map[string][]string
}

type NewDecodeRequestBodyErrorInput struct {
	Code        string
	Message     string
	FieldErrors map[string][]string
}

func NewDecodeRequestBodyError(input NewDecodeRequestBodyErrorInput) *DecodeRequestBodyError {
	return &DecodeRequestBodyError{
		Code:        input.Code,
		Message:     input.Message,
		FieldErrors: input.FieldErrors,
	}
}
func (e DecodeRequestBodyError) Error() string {
	if e.Message != "" {
		return e.Message
	}

	return "error decoding request body"
}

func (e DecodeRequestBodyError) IsEmpty() bool {
	return e.Code == "" && e.Message == "" && len(e.FieldErrors) == 0
}
