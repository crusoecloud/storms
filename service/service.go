package service

import (
	"crypto/tls"
	"fmt"
	"net"

	uuid "github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"gitlab.com/crusoeenergy/island/storage/storms/client"
	"gitlab.com/crusoeenergy/island/storage/storms/configs"
	"gitlab.com/crusoeenergy/island/storage/storms/model"
	"gitlab.com/crusoeenergy/island/storage/storms/resource"
	"gitlab.com/crusoeenergy/schemas/api/island/v2/storagemulti"
	grpcutil "gitlab.com/crusoeenergy/schemas/utils/rpc/grpc"
)

// --- Begin clientMap

// Enables identifying generic client via global uuidtype.
type clientMap struct {
	clients map[string]client.Client
}

func newClientMap() *clientMap {
	return &clientMap{
		clients: make(map[string]client.Client),
	}
}

func (cm *clientMap) addClient(c client.Client) string {
	_uuid := uuid.NewString()
	cm.clients[_uuid] = c

	return _uuid
}

//nolint:unused,ireturn // will be used in the future; return interface to manage generic
func (cm *clientMap) getClient(_uuid string) client.Client {
	val, ok := cm.clients[_uuid]
	if ok {
		return val
	}

	return nil
}

func (cm *clientMap) getAllClients() []client.Client {
	return lo.Values(cm.clients)
}

// --- Begin resourceMap

// Enables identifying resource via global uuid.
type resourceMap struct {
	resources map[string]resource.Resource
}

func newResourceMap() *resourceMap {
	return &resourceMap{
		resources: make(map[string]resource.Resource),
	}
}

func (rm *resourceMap) addResource(r resource.Resource) string {
	_uuid := uuid.NewString()
	rm.resources[_uuid] = r

	return _uuid
}

//nolint:unused,ireturn // will be used in the future; return interface to manage generic
func (rm *resourceMap) getResource(_uuid string) resource.Resource {
	val, ok := rm.resources[_uuid]
	if ok {
		return val
	}

	return nil
}

// --- Begin Service

type Service struct {
	// ctx      context.Context
	listener net.Listener
	endpoint string

	clients   *clientMap   // TODO - we may want this in a file
	resources *resourceMap // TODO - we may want this in a file

	*grpc.Server
	storagemulti.UnimplementedStormsServiceServer
}

func NewService(endpoint string) *Service {
	s := &Service{
		// ctx:       ctx,
		clients:   newClientMap(),
		resources: newResourceMap(),
		endpoint:  endpoint,
	}

	return s
}

func (s *Service) Start() error {
	err := s.loadClusterConfig()
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
func (s *Service) loadClusterConfig() error {
	clusterConfig, err := model.LoadClusterConfig(configs.Get().ClusterFile)
	if err != nil {
		return fmt.Errorf("failed to load cluster config: %w", err)
	}

	for _, cluster := range clusterConfig.Clusters {
		log.Info().Msgf("TODO: create client with cluster endpoint: %s\n", cluster.Endpoint)
		clients := []client.Client{}
		for _, client := range clients {
			s.clients.addClient(client)
		}
	}

	return nil
}

// Creates in-memory global metadata mapping of resource
//
//nolint:unparam // always returning nil, but this will change after TODO is addressed
func (s *Service) fetchClusterMetadata() error {
	for _, client := range s.clients.getAllClients() {
		log.Info().Msgf("TODO: fetch all the resources from client %v\n", client)
		resources := []resource.Resource{}
		for _, resource := range resources {
			s.resources.addResource(resource)
		}
	}

	return nil
}

// Registers services and serves.
func (s *Service) serve() error {
	var tlsConfig *tls.Config

	s.Server = grpcutil.NewServer(tlsConfig)

	listener, err := net.Listen("tcp", s.endpoint)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", s.endpoint, err)
	}
	s.listener = listener

	reflection.Register(s)
	storagemulti.RegisterStormsServiceServer(s, s)

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
