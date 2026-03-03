package errors

type ApiError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	StatusCode int    `json:"-"`
	Err        error  `json:"-"`
}

func (a ApiError) Error() string {
	return a.Message
}
