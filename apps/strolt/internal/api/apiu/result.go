package apiu

type ResultError struct {
	Error string `json:"error"`
}

type ResultSuccess struct {
	Data string `json:"data"`
}
