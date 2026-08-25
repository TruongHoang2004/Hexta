package service

type baseService struct {
}

func NewBaseService() *baseService {
	return &baseService{}
}

// func (b *baseService) grpToIError(ctx context.Context, inputErr error) *errors.Error {
// 	var ierr errors.Error
// 	grpcErr, ok := status.FromError(inputErr)
// 	if !ok {
// 		return errors.ErrSystemError(ctx, fmt.Sprintf("grpc error convert failed, err:[%s]", inputErr.Error()))
// 	}

// 	err := json.Unmarshal([]byte(grpcErr.Message()), &ierr)
// 	if err != nil {
// 		return errors.ErrSystemError(ctx, fmt.Sprintf("grpc error unmarshal failed with input [%s]", inputErr.Error()))
// 	}

// 	return &ierr
// }
