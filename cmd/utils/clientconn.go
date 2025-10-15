package utils

import (
	"context"
	"fmt"
	"io"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	admin "gitlab.com/crusoeenergy/island/storage/storms/api/gen/go/admin/v1"
	storms "gitlab.com/crusoeenergy/island/storage/storms/api/gen/go/storms/v1"
)

const grpcAddr = "127.0.0.1:9290"

// Create a Storage Management Service client.
func CreateStorMSClient() (storms.StorageManagementServiceClient, *grpc.ClientConn, error) {
	// Dial gRPC server with modern options
	conn, err := grpc.NewClient(
		grpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()), // insecure; replace with TLS in prod
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create grpc client: %w", err)
	}

	client := storms.NewStorageManagementServiceClient(conn)

	return client, conn, nil
}

// Create a Storage Management Service client.
func CreateStorMSClientV2() (storms.StorageManagementServiceClient, *grpc.ClientConn, error) {
	// Dial gRPC server with modern options
	conn, err := grpc.NewClient(
		grpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()), // insecure; replace with TLS in prod
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create grpc client: %w", err)
	}

	client := storms.NewStorageManagementServiceClient(conn)

	return client, conn, nil
}

type StorMSClientProvider func(context.Context) (storms.StorageManagementServiceClient, io.Closer, error)

func DefaultStorMSProvider(_ context.Context) (storms.StorageManagementServiceClient, io.Closer, error) {
	return CreateStorMSClient()
}

func CreateAdminClient() (admin.AdminServiceClient, *grpc.ClientConn, error) {
	// Dial gRPC server with modern options
	conn, err := grpc.NewClient(
		grpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()), // insecure; replace with TLS in prod
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create grpc client: %w", err)
	}

	client := admin.NewAdminServiceClient(conn)

	return client, conn, nil
}

type AdminClientProvider func(context.Context) (admin.AdminServiceClient, io.Closer, error)

func DefaultAdminProvider(_ context.Context) (admin.AdminServiceClient, io.Closer, error) {
	return CreateAdminClient()
}
