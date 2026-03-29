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

type UserClient struct {
	client pb.UserServiceClient
	conn   *grpc.ClientConn
}

var userClient *UserClient

func NewUserClient() (*UserClient, error) {
	if userClient != nil {
		return userClient, nil
	}

	url := config.AppConfig.Service.User
	if url == "" {
		return nil, fmt.Errorf("user service url is not configured")
	}

	// Create a new gRPC client connection using the address from config
	// We use insecure credentials for internal communication, assuming secured network or dev environment
	// In production, you might want to use TLS
	conn, err := grpc.NewClient(
		url,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create user service client: %w", err)
	}

	client := pb.NewUserServiceClient(conn)

	return &UserClient{
		client: client,
		conn:   conn,
	}, nil
}

// Close closes the underlying gRPC connection
func (c *UserClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// CreateUser calls the CreateUser RPC
func (c *UserClient) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return c.client.CreateUser(ctx, req)
}

// GetUser calls the GetUser RPC
func (c *UserClient) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return c.client.GetUser(ctx, req)
}

// GetUsers calls the GetUsers RPC
func (c *UserClient) GetUsers(ctx context.Context, req *pb.GetUsersRequest) (*pb.GetUsersResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return c.client.GetUsers(ctx, req)
}

// ListUsers calls the ListUsers RPC
func (c *UserClient) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return c.client.ListUsers(ctx, req)
}

// DeleteUser calls the DeleteUser RPC
func (c *UserClient) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return c.client.DeleteUser(ctx, req)
}
