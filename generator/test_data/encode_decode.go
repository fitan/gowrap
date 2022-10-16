package test_data

type Response struct {
	Success bool        `json:"success"`
	Code    int         `json:"code"`
	Data    interface{} `json:"data,omitempty"`
	Error   error       `json:"error,omitempty" swaggertype:"string"`
	Message string      `json:"message"`
	TraceId string      `json:"traceId"`
}
