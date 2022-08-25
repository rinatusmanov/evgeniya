package types

// ErrorData used like rpc field error.data in response with error.
// It will be showed in openapi spec if you passed it in service description.
type ErrorData struct {
	ClientMessage string `json:"client_message"`
}
