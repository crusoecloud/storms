package service

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	admin "gitlab.com/crusoeenergy/island/storage/storms/api/gen/go/admin/v1"
	storms "gitlab.com/crusoeenergy/island/storage/storms/api/gen/go/storms/v1"
	"gitlab.com/crusoeenergy/island/storage/storms/client"
	"gitlab.com/crusoeenergy/island/storage/storms/client/models"
	appconfigs "gitlab.com/crusoeenergy/island/storage/storms/internal/app/configs"
	alloc "gitlab.com/crusoeenergy/island/storage/storms/internal/service/allocator"
	cluster "gitlab.com/crusoeenergy/island/storage/storms/internal/service/cluster"
	serviceconfigs "gitlab.com/crusoeenergy/island/storage/storms/internal/service/configs"
	resource "gitlab.com/crusoeenergy/island/storage/storms/internal/service/resource"
	translator "gitlab.com/crusoeenergy/island/storage/storms/internal/service/translator"
)

const (
	requestTimeoutMin = 1
	tcpProtocol       = "tcp"
)

type clientTranslator interface {
	AttachVolume(ctx context.Context, c client.Client, req *storms.AttachVolumeRequest,
	) (*storms.AttachVolumeResponse, error)
	CreateSnapshot(ctx context.Context, c client.Client, req *storms.CreateSnapshotRequest,
	) (*storms.CreateSnapshotResponse, error)
	CreateVolume(ctx context.Context, c client.Client, req *storms.CreateVolumeRequest,
	) (*storms.CreateVolumeResponse, error)
	DeleteSnapshot(ctx context.Context, c client.Client, req *storms.DeleteSnapshotRequest,
	) (*storms.DeleteSnapshotResponse, error)
	DeleteVolume(ctx context.Context, c client.Client, req *storms.DeleteVolumeRequest,
	) (*storms.DeleteVolumeResponse, error)
	DetachVolume(ctx context.Context, c client.Client, req *storms.DetachVolumeRequest,
	) (*storms.DetachVolumeResponse, error)
	GetSnapshot(ctx context.Context, c client.Client, req *storms.GetSnapshotRequest,
	) (*storms.GetSnapshotResponse, error)
	GetSnapshots(ctx context.Context, c client.Client, _ *storms.GetSnapshotsRequest,
	) (*storms.GetSnapshotsResponse, error)
	GetVolume(ctx context.Context, c client.Client, req *storms.GetVolumeRequest,
	) (*storms.GetVolumeResponse, error)
	GetVolumes(ctx context.Context, c client.Client, _ *storms.GetVolumesRequest,
	) (*storms.GetVolumesResponse, error)
	ResizeVolume(ctx context.Context, c client.Client, req *storms.ResizeVolumeRequest,
	) (*storms.ResizeVolumeResponse, error)
}

// ClusterManager manages the lifecycle of clients for storage clusters.
// It is the source of truth for which clusters are currently managed.
type clusterManager interface {
	Set(clusterID string, client client.Client) error
	Remove(clusterID string) error
	Get(clusterID string) (client.Client, error)
	AllIDs() []string
	Count() int
}

// // resourceManager maps individual storage resources to the cluster that owns them.
// type resourceManager interface {
// 	Map(resourceID, clusterID string) error
// 	Unmap(resourceID string) error
// 	OwnerCluster(resourceID string) (string, error)
// 	ResourceCount() int
// 	ResourceCountByCluster(clusterID string) int
// 	GetAllClusterResources() map[string][]string
// }

type resourceManager interface {
	Map(r *resource.Resource) error
	Unmap(resourceID string) error
	GetResourceCluster(resourceID string) (string, error)
	GetResourceCount() int
	GetResourcesOfCluster(clusterID string) []*resource.Resource
	GetResourcesOfAllClusters() map[string][]*resource.Resource
}

// Allocator decides which cluster a new resource should be placed on.
type allocator interface {
	SelectClusterForNewResource() (string, error) // Returns clusterID
}

type Service struct {
	// Stores configuration of clusters under management
	clusterConfigs *serviceconfigs.ClusterConfig

	// Translates service models to client interface models.
	clientTranslator clientTranslator

	// Components for caching resource and cluster metadata. Routes requests.
	clusterManager  clusterManager
	resourceManager resourceManager
	// resourceManager resourceManager
	allocator allocator

	// Components for creating gRPC server and service
	listener net.Listener
	endpoint string
	*grpc.Server
	storms.UnimplementedStorageManagementServiceServer
	admin.UnimplementedAdminServiceServer
}

func NewService(endpoint string) *Service {
	clusterManger := cluster.NewInMemoryManager()
	s := &Service{
		endpoint:         endpoint,
		clientTranslator: translator.NewClientTranslator(),
		clusterManager:   clusterManger,
		resourceManager:  resource.NewInMemoryManager(),
		allocator:        alloc.NewRoundRobinAllocator(clusterManger),
	}

	return s
}

// Loads cluster configuration, fetches resource metadata from each cluster, then serves.
func (s *Service) Start() error {
	err := s.loadClusterConfigs()
	if err != nil {
		return fmt.Errorf("failed to load cluster configuration: %w", err)
	}

	s.syncClusterManager()
	s.syncResourceManager()

	err = s.serve()
	if err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}

// Load cluster configuration from configuration file.
func (s *Service) loadClusterConfigs() error {
	clusterConfigs, err := serviceconfigs.LoadClusterConfig(appconfigs.Get().ClusterFile)
	if err != nil {
		return fmt.Errorf("failed to load cluster config: %w", err)
	}

	s.clusterConfigs = clusterConfigs
	log.Info().Msgf("Loaded cluster config from file: %s", appconfigs.Get().ClusterFile)

	return nil
}

// Adds clients to back the clusters to be managed, specified by cluster configuration.
// Removes any clients that are not specified in the cluster configuration.
func (s *Service) syncClusterManager() {
	clusters := lo.SliceToMap(
		s.clusterConfigs.Clusters,
		func(c serviceconfigs.Cluster) (string, *serviceconfigs.Cluster) {
			return c.ClusterID, &c
		})

	// Set cluster-client pairings.
	for clusterID, cluster := range clusters {
		c, err := client.NewClient(cluster.Vendor, cluster.VendorConfig)
		if err != nil {
			log.Err(err).Str("cluster_id", clusterID).Msg("failed to create new client for cluster")
		}

		err = s.clusterManager.Set(clusterID, c)
		if err != nil {
			log.Err(err).Str("cluster_id", clusterID).Msg("failed to add client to cluster manager")
		}
	}

	// Remove cluster-client pairings not specified in configuration.
	managedClusterIDs := s.clusterManager.AllIDs()
	desiredClusterIDs := lo.Keys(clusters)
	for _, managedClusterID := range managedClusterIDs {
		if !lo.Contains(desiredClusterIDs, managedClusterID) {
			err := s.clusterManager.Remove(managedClusterID)
			if err != nil {
				log.Warn().Str("cluster_id", managedClusterID).Interface("err", err).Msg("failed to remove cluster")
			}
		}
	}
	log.Info().Msg("Synced Cluster Manager.")
}

// Adds all resources from managed clusters into resource mapper.
func (s *Service) syncResourceManager() {
	// Map resources for all clusters that are specified in the configuration.
	wg := sync.WaitGroup{}
	managedClusterIDs := s.clusterManager.AllIDs()
	for _, clusterID := range managedClusterIDs {
		wg.Add(1)
		go func(cid string) {
			defer wg.Done()
			resources := s.fetchResourcesFromCluster(cid)
			for _, r := range resources {
				err := s.resourceManager.Map(r)
				if err != nil {
					log.Warn().Str("resource_id", r.ID).Interface("err", err).Msg("failed to map resource")
				}
			}
		}(clusterID)
	}
	wg.Wait()

	// Unmap resources with clusters that are not specified in the configuration.
	for clusterID, resources := range s.resourceManager.GetResourcesOfAllClusters() {
		unmanaged := !lo.Contains(managedClusterIDs, clusterID)
		if unmanaged {
			for _, r := range resources {
				err := s.resourceManager.Unmap(r.ID)
				if err != nil {
					log.Warn().Str("resource_id", r.ID).Interface("err", err).Msg("failed to unmap resource")
				}
			}
		}
	}

	log.Info().Msg("Synced Resource Manager.")
}

func (s *Service) fetchResourcesFromCluster(clusterID string) []*resource.Resource {
	c, err := s.clusterManager.Get(clusterID)
	if err != nil {
		log.Err(err).Str("cluster_id", clusterID).Msg("TODO")
	}

	// Set up timeout.
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeoutMin*time.Minute)
	defer cancel()

	resources := make([]*resource.Resource, 0)

	// Fetch volumes.
	getVolResp, err := c.GetVolumes(ctx, &models.GetVolumesRequest{})
	if err != nil {
		log.Err(err).Str("cluster_id", clusterID).Msg("failed to get volumes")
	}
	volumes := lo.Map(getVolResp.Volumes, func(v *models.Volume, _ int) *resource.Resource {
		return &resource.Resource{
			ID:           v.UUID,
			ClusterID:    clusterID,
			ResourceType: resource.TypeVolume,
		}
	})
	resources = append(resources, volumes...)
	log.Info().Str("cluster_id", clusterID).Msgf("fetched %d volumes", len(volumes))

	// Fetch snapshots.
	getSnapshotResp, err := c.GetSnapshots(ctx, &models.GetSnapshotsRequest{})
	if err != nil {
		log.Err(err).Str("cluster_id", clusterID).Msg("failed to get snaphots")
	}
	snapshots := lo.Map(getSnapshotResp.Snapshots, func(s *models.Snapshot, _ int) *resource.Resource {
		return &resource.Resource{
			ID:           s.UUID,
			ClusterID:    clusterID,
			ResourceType: resource.TypeSnapshot,
		}
	})
	resources = append(resources, snapshots...)
	log.Info().Str("cluster_id", clusterID).Msgf("fetched %d snapshots", len(snapshots))

	return resources
}

// Registers services and serves.
func (s *Service) serve() error {
	s.Server = grpc.NewServer(
		grpc.UnaryInterceptor(loggingUnaryInterceptor),
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
	admin.RegisterAdminServiceServer(s, s)

	log.Info().Msg("Starting service")
	go func() {
		if err := s.Serve(s.listener); err != nil {
			log.Err(fmt.Errorf("failed to Start service: %w", err))
		}
	}()

	return nil
}

func (s *Service) getClientForResource(resourceID string) (string, client.Client, error) {
	clusterID, err := s.resourceManager.GetResourceCluster(resourceID)
	if err != nil {
		return "", nil, fmt.Errorf("failed to get cluster for resource: %w", err)
	}
	c, err := s.clusterManager.Get(clusterID)
	if err != nil {
		return "", nil, fmt.Errorf("failed to get client for cluster: %w", err)
	}

	return clusterID, c, nil
}

func (s *Service) Stop() error {
	s.Server.GracefulStop()

	return nil
}

// Begin -- static functions

func loggingUnaryInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	start := time.Now()
	requestLogger := log.Info()
	requestLogger.
		Str("grpc_method", info.FullMethod).
		Interface("request", req).
		Msg("Unary request received")

	resp, err = handler(ctx, req)
	duration := time.Since(start)

	// Extract gRPC status code from the error.
	st, _ := status.FromError(err)
	responseLogger := log.Info() // Default to Info level for success.
	if err != nil {
		responseLogger = log.Error().Err(err) // Use Error level if an error occurred.
	}

	// Log the response details.
	responseLogger.
		Str("grpc_method", info.FullMethod).
		Stringer("grpc_code", st.Code()). // Log the string representation of the code (e.g., "OK", "NotFound").
		Dur("duration", duration).
		Msg("Unary response sent")

	return resp, err
}

func loggingStreamInterceptor(
	srv interface{},
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	start := time.Now()
	requestLogger := log.Info()
	requestLogger.Msgf("gRPC stream started: %s", info.FullMethod)

	err := handler(srv, ss)
	duration := time.Since(start)

	// Extract gRPC status code from the error.
	st, _ := status.FromError(err)

	// Determine the log level based on the error.
	responseLogger := log.Info() // Default to Info level for success.
	if err != nil {
		responseLogger = log.Error().Err(err) // Use Error level if an error occurred.
	}

	// Log the final status of the stream.
	responseLogger.
		Str("grpc_method", info.FullMethod).
		Stringer("grpc_code", st.Code()).
		Dur("duration", duration).
		Msg("Stream finished")

	return err
}
