package response

type Response[T any] struct {
	Data T           `json:"data"`
	Meta interface{} `json:"meta"`
}

func NewSuccessResponse[T any](data T, meta interface{}) *Response[T] {
	return &Response[T]{
		Data: data,
		Meta: meta,
	}
}
