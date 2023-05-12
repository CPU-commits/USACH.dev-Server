package res

type Response struct {
	Message string                 `json:"message,omitempty" extensions:"x-omitempty" example:"Error message"`
	Data    map[string]interface{} `json:"body,omitempty" extensions:"x-omitempty"`
}

type ErrorRes struct {
	Err        error
	StatusCode int
}
