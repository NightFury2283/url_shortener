package response

type Response struct {
	Status string      `json:"status"`           // "OK" or "Error"
	Error  string      `json:"error,omitempty"`  // omitempty to not return empty field
}


const (
	StatusOK    = "OK"
	StatusError = "Error"
)

func OK() Response {
	return Response{
		Status: StatusOK,
	}
}

func Error(errMsg string) Response {
	return Response{
		Status: StatusError,
		Error:  errMsg,
	}
}