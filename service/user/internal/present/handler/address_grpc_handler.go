package handler

import (
	"context"

	"gitlab.com/ecommercehub1/user/internal/core/model"
	"gitlab.com/ecommercehub1/user/internal/core/service"
	pb "gitlab.com/ecommercehub1/user/proto"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AddressGrpcHandler struct {
	pb.UnimplementedAddressServiceServer
	service *service.AddressService
}

func NewAddressGrpcHandler(service *service.AddressService) *AddressGrpcHandler {
	return &AddressGrpcHandler{
		service: service,
	}
}

func (h *AddressGrpcHandler) Register(s *grpc.Server) {
	pb.RegisterAddressServiceServer(s, h)
}

func (h *AddressGrpcHandler) CreateAddress(ctx context.Context, req *pb.CreateAddressRequest) (*pb.CreateAddressResponse, error) {
	address := &model.Address{
		UserID:    req.UserId,
		Receiver:  req.Receiver,
		Phone:     req.Phone,
		City:      req.City,
		District:  req.District,
		Ward:      req.Ward,
		Street:    req.Street,
		Details:   req.Details,
		IsDefault: req.IsDefault,
	}

	if err := h.service.CreateAddress(address); err != nil {
		return nil, err
	}

	return &pb.CreateAddressResponse{
		Address: mapAddressToProto(address),
	}, nil
}

func (h *AddressGrpcHandler) UpdateAddress(ctx context.Context, req *pb.UpdateAddressRequest) (*pb.UpdateAddressResponse, error) {
	address, err := h.service.GetAddressByID(req.Id)
	if err != nil {
		return nil, err
	}

	// Update fields
	address.Receiver = req.Receiver
	address.Phone = req.Phone
	address.City = req.City
	address.District = req.District
	address.Ward = req.Ward
	address.Street = req.Street
	address.Details = req.Details
	address.IsDefault = req.IsDefault

	if err := h.service.UpdateAddress(address); err != nil {
		return nil, err
	}

	return &pb.UpdateAddressResponse{
		Address: mapAddressToProto(address),
	}, nil
}

func (h *AddressGrpcHandler) DeleteAddress(ctx context.Context, req *pb.DeleteAddressRequest) (*pb.DeleteAddressResponse, error) {
	if err := h.service.DeleteAddress(req.Id); err != nil {
		return nil, err
	}
	return &pb.DeleteAddressResponse{
		Success: true,
	}, nil
}

func (h *AddressGrpcHandler) GetAddress(ctx context.Context, req *pb.GetAddressRequest) (*pb.GetAddressResponse, error) {
	address, err := h.service.GetAddressByID(req.Id)
	if err != nil {
		return nil, err
	}
	return &pb.GetAddressResponse{
		Address: mapAddressToProto(address),
	}, nil
}

func (h *AddressGrpcHandler) ListAddresses(ctx context.Context, req *pb.ListAddressesRequest) (*pb.ListAddressesResponse, error) {
	addresses, total, err := h.service.ListUserAddresses(req.UserId, req.Page, req.Limit)
	if err != nil {
		return nil, err
	}

	var pbAddresses []*pb.Address
	for _, addr := range addresses {
		pbAddresses = append(pbAddresses, mapAddressToProto(addr))
	}

	return &pb.ListAddressesResponse{
		Addresses: pbAddresses,
		Total:     int32(total),
	}, nil
}

func mapAddressToProto(address *model.Address) *pb.Address {
	return &pb.Address{
		Id:        address.ID,
		UserId:    address.UserID,
		Receiver:  address.Receiver,
		Phone:     address.Phone,
		City:      address.City,
		District:  address.District,
		Ward:      address.Ward,
		Street:    address.Street,
		Details:   address.Details,
		IsDefault: address.IsDefault,
		CreatedAt: timestamppb.New(address.CreatedAt),
		UpdatedAt: timestamppb.New(address.UpdatedAt),
	}
}
