package utils

import (
	"context"
	"fmt"
	"io"

	admin "gitlab.com/crusoeenergy/island/storage/storms/pkg/api/gen/go/admin/v1"
	storms "gitlab.com/crusoeenergy/island/storage/storms/pkg/api/gen/go/storms/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type CmdFactory struct {
	TargetAddr string

	AdminClientProvider  AdminClientProvider
	StorMSClientProvider StorMSClientProvider
}

func NewCmdFactory() *CmdFactory {
	f := &CmdFactory{}
	f.AdminClientProvider = f.DefaultAdminProvider
	f.StorMSClientProvider = f.DefaultStorMSProvider

	return f
}

// Create a Storage Management Service client.
func (f *CmdFactory) CreateStorMSClient() (storms.StorageManagementServiceClient, *grpc.ClientConn, error) {
	// Dial gRPC server with modern options
	conn, err := grpc.NewClient(
		f.TargetAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()), // insecure; replace with TLS in prod
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create grpc client: %w", err)
	}

	client := storms.NewStorageManagementServiceClient(conn)

	return client, conn, nil
}

type StorMSClientProvider func(context.Context) (storms.StorageManagementServiceClient, io.Closer, error)

func (f *CmdFactory) DefaultStorMSProvider(_ context.Context,
) (storms.StorageManagementServiceClient, io.Closer, error) {
	return f.CreateStorMSClient()
}

func (f *CmdFactory) CreateAdminClient() (admin.AdminServiceClient, *grpc.ClientConn, error) {
	// Dial gRPC server with modern options
	conn, err := grpc.NewClient(
		f.TargetAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()), // insecure; replace with TLS in prod
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create grpc client: %w", err)
	}

	client := admin.NewAdminServiceClient(conn)

	return client, conn, nil
}

type AdminClientProvider func(context.Context) (admin.AdminServiceClient, io.Closer, error)

func (f *CmdFactory) DefaultAdminProvider(_ context.Context) (admin.AdminServiceClient, io.Closer, error) {
	return f.CreateAdminClient()
}
