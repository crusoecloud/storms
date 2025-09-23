package service

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	storms "gitlab.com/crusoeenergy/island/storage/storms/api/gen/go/storms/v1"
	"gitlab.com/crusoeenergy/island/storage/storms/client"
	appconfigs "gitlab.com/crusoeenergy/island/storage/storms/internal/app/configs"
	serviceconfigs "gitlab.com/crusoeenergy/island/storage/storms/internal/service/configs"
)

const tcpProtocol = "tcp"

var (
	errNoClients           = errors.New("no clients")
	errUnmappedResource    = errors.New("failed to map resource to client")
	errUnmappedClientID    = errors.New("unmapped client ID")
	errDuplicateClientID   = errors.New("duplicate client IDs")
	errDuplicateResourceID = errors.New("duplicate resource IDs")
)

type Service struct {
	listener net.Listener
	endpoint string

	ResourceManager  *ResourceManager
	ClientTranslator *ClientTranslator

	*grpc.Server
	storms.UnimplementedStorageManagementServiceServer
}

func NewService(endpoint string) *Service {
	s := &Service{
		endpoint:         endpoint,
		ResourceManager:  newResourceManager(RoundRobinClientAllocAlgo),
		ClientTranslator: NewClientTranslator(),
	}

	return s
}

// Loads cluster configuration, fetches resource metadata from each cluster, then serves.
func (s *Service) Start() error {
	err := s.loadClusterConfigs()
	if err != nil {
		return fmt.Errorf("failed to load cluster configuration: %w", err)
	}

	err = s.fetchClusterMetadata()
	if err != nil {
		return fmt.Errorf("failed to fetch cluster metadata: %w", err)
	}

	err = s.serve()
	if err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}

// Creates generic client connections and create in-memory mappying of client uuid to the client.
func (s *Service) loadClusterConfigs() error {
	log.Info().Msgf("Loading cluster config from file: %s", appconfigs.Get().ClusterFile)
	clusterConfigs, err := serviceconfigs.LoadClusterConfig(appconfigs.Get().ClusterFile) // TODO - refactor
	if err != nil {
		return fmt.Errorf("failed to load cluster config: %w", err)
	}

	for _, cluster := range clusterConfigs.Clusters {
		clusterID := cluster.ClusterID
		c, err := client.NewClient(cluster.Vendor, cluster.VendorConfig)
		if err != nil {
			return fmt.Errorf("failed to create new %s client: %w", cluster.Vendor, err)
		}

		err = s.ResourceManager.addClient(clusterID, c)
		if err != nil {
			return fmt.Errorf("failed to add client to resource manager: %w", err)
		}
	}
	log.Info().Msgf("StorMS configured")

	return nil
}

// Creates in-memory global metadata mapping of resource.
func (s *Service) fetchClusterMetadata() error {
	log.Info().Msgf("Fetching metadata of clusters")

	for _, clientID := range s.ResourceManager.getAllClientIDs() {
		resourceIDs, err := s.ResourceManager.fetchAllResourcesFromClient(clientID)
		log.Info().Msgf("[ClientID=%s] - Found %d resources", clientID, len(resourceIDs))
		if err != nil {
			return fmt.Errorf("failed to fetch all resources from client: %w", err)
		}

		for _, resourceID := range resourceIDs {
			err := s.ResourceManager.addResource(resourceID, clientID)
			if err != nil {
				return fmt.Errorf("failed to add resource: %w", err)
			}
		}
	}

	s.logClusterMetadata()

	return nil
}

// Log cluster metadata.
func (s *Service) logClusterMetadata() {
	log.Info().Msg(
		"Fetched metadata of all clusters",
	)
	log.Info().Msgf(
		"Managing %d clusters with %d total resources",
		len(s.ResourceManager.getAllClientIDs()),
		len(lo.Keys(s.ResourceManager.resourceClientMap)),
	)
}

// Registers services and serves.
func (s *Service) serve() error {
	s.Server = grpc.NewServer(grpc.UnaryInterceptor(loggingUnaryInterceptor),
		grpc.StreamInterceptor(loggingStreamInterceptor),
	)

	listenConfig := net.ListenConfig{}
	listener, err := listenConfig.Listen(context.Background(), tcpProtocol, s.endpoint)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", s.endpoint, err)
	}
	s.listener = listener

	reflection.Register(s)
	storms.RegisterStorageManagementServiceServer(s, s)

	log.Info().Msg("Starting service")
	go func() {
		if err := s.Serve(s.listener); err != nil {
			log.Err(fmt.Errorf("failed to Start service: %w", err))
		}
	}()

	return nil
}

func (s *Service) Stop() error {
	// Close gRPC connections

	return nil
}

// Begin -- static functions

func loggingUnaryInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	log.Info().Msgf("gRPC method: %s, req=%v", info.FullMethod, req)
	resp, err = handler(ctx, req)

	return resp, err
}

func loggingStreamInterceptor(
	srv interface{},
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	log.Info().Msgf("gRPC stream started: %s", info.FullMethod)
	err := handler(srv, ss)

	return err
}
