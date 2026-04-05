package client

import (
	"context"
	"fmt"
	"time"

	"gitlab.com/ecommercehub1/api/config"
	pb "gitlab.com/ecommercehub1/user/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type MerchantClient struct {
	client pb.MerchantServiceClient
	conn   *grpc.ClientConn
}

var merchantClient *MerchantClient

func NewMerchantClient() (*MerchantClient, error) {
	if merchantClient != nil {
		return merchantClient, nil
	}

	url := config.AppConfig.Service.Merchant
	if url == "" {
		return nil, fmt.Errorf("merchant service url is not configured")
	}

	conn, err := grpc.NewClient(
		url,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create merchant service client: %w", err)
	}

	merchantClient = &MerchantClient{
		client: pb.NewMerchantServiceClient(conn),
		conn:   conn,
	}
	return merchantClient, nil
}

// Close closes the underlying gRPC connection
func (c *MerchantClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// CreateMerchant calls the CreateMerchant RPC
func (c *MerchantClient) CreateMerchant(ctx context.Context, req *pb.CreateMerchantRequest) (*pb.CreateMerchantResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return c.client.CreateMerchant(ctx, req)
}

// GetMerchant calls the GetMerchant RPC
func (c *MerchantClient) GetMerchant(ctx context.Context, req *pb.GetMerchantRequest) (*pb.GetMerchantResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return c.client.GetMerchant(ctx, req)
}

// GetMerchantByOwner calls the GetMerchantByOwner RPC
func (c *MerchantClient) GetMerchantByOwner(ctx context.Context, req *pb.GetMerchantByOwnerRequest) (*pb.GetMerchantByOwnerResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return c.client.GetMerchantByOwner(ctx, req)
}

// UpdateMerchant calls the UpdateMerchant RPC
func (c *MerchantClient) UpdateMerchant(ctx context.Context, req *pb.UpdateMerchantRequest) (*pb.UpdateMerchantResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return c.client.UpdateMerchant(ctx, req)
}

// ListMerchants calls the ListMerchants RPC
func (c *MerchantClient) ListMerchants(ctx context.Context, req *pb.ListMerchantsRequest) (*pb.ListMerchantsResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return c.client.ListMerchants(ctx, req)
}

// DeleteMerchant calls the DeleteMerchant RPC
func (c *MerchantClient) DeleteMerchant(ctx context.Context, req *pb.DeleteMerchantRequest) (*pb.DeleteMerchantResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return c.client.DeleteMerchant(ctx, req)
}
