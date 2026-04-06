package models

type ApiError struct {
	Code    string              `json:"code"`
	Message string              `json:"message"`
	Fields  map[string][]string `json:"fields,omitempty"`
}

type ApiErrorResponse struct {
	Error ApiError `json:"error"`
}

type ApiEntitiesMeta struct {
	Page       int `json:"page"`
	PageSize   int `json:"pageSize"`
	TotalItems int `json:"totalItems"`
	TotalPages int `json:"totalPages"`
}

type ApiSingleResponse[T any] struct {
	Data T `json:"data"`
}

type ApiEntitiesResponse[T any] struct {
	Data []T             `json:"data"`
	Meta ApiEntitiesMeta `json:"meta"`
}
