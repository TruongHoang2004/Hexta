package response

type Response struct {
	Data interface{} `json:"data"`
	Meta interface{} `json:"meta"`
}

func NewSuccessResponse(data interface{}, meta interface{}) interface{} {
	return &Response{
		Data: data,
		Meta: meta,
	}
}
