package models

type DecodeRequestBodyError struct {
	Code        string
	Message     string
	FieldErrors map[string][]string
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
