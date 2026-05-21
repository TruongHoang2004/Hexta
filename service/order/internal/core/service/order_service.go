package service

import (
	"context"

	"github.com/oklog/ulid/v2"
	"gitlab.com/ecommercehub1/lib/pkg/errors"
	"gitlab.com/ecommercehub1/order/internal/core/model"
	"gitlab.com/ecommercehub1/order/internal/infrastructure/kafka"
	"gitlab.com/ecommercehub1/order/internal/present/http/dto"
	"gitlab.com/ecommercehub1/order/internal/repository"
)

type IOrderService interface {
	CreateOrder(ctx context.Context, req *dto.CreateOrderRequest) (*dto.OrderResponse, *errors.Error)
}

type OrderService struct {
	orderRepo repository.IOrderRepository
	producer  kafka.Producer
}

func NewOrderService(orderRepo repository.IOrderRepository, producer kafka.Producer) *OrderService {
	return &OrderService{
		orderRepo: orderRepo,
		producer:  producer,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, req *dto.CreateOrderRequest) (*dto.OrderResponse, *errors.Error) {
	// Generate an ID for the new order
	orderID := ulid.Make().String()

	// In a high-throughput scenario, instead of writing to DB immediately,
	// we publish an OrderPlaced event.
	order := &model.Order{
		ID:        orderID,
		UserID:    req.UserID,
		ProductID: req.ProductID,
		Quantity:  req.Quantity,
		Status:    "PENDING",
	}

	// Publish to Kafka
	err := s.producer.Publish(ctx, "orders", orderID, order)
	if err != nil {
		return nil, errors.ErrSystemError(ctx, "Failed to publish order event")
	}

	// Return response immediately
	return &dto.OrderResponse{
		ID:        orderID,
		Status:    "PENDING",
	}, nil
}
