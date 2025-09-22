package service

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"

	"gitlab.com/crusoeenergy/island/storage/storms/client"
	"gitlab.com/crusoeenergy/island/storage/storms/client/models"
)

func newResourceManager() *ResourceManager {
	return &ResourceManager{
		clients:           make(map[string]client.Client),
		resourceClientMap: map[string]string{},
	}
}

type ResourceManager struct {
	clients           map[string]client.Client // key:value::clientID:clientInstance
	resourceClientMap map[string]string        // key:value::resourceID:clientID
}

// Adds client into ResourceManager and makes it fetchable with uuid.
func (r *ResourceManager) addClient(clusterID string, a client.Client) error {
	_, ok := r.clients[clusterID]
	if ok {
		return errDuplicateClientID
	}
	r.clients[clusterID] = a

	return nil
}

//nolint:ireturn // returning interface to support generic type
func (r *ResourceManager) allocateClient() (string, client.Client, error) {
	// TODO - using client allocation algorithm instead
	clientIDs := lo.Keys(r.clients)
	if len(clientIDs) == 0 {
		return "", nil, errNoClients
	}

	return clientIDs[0], r.clients[clientIDs[0]], nil
}

//nolint:ireturn // returning interface to support generic type
func (r *ResourceManager) getClientForResource(resourceID string) (client.Client, error) {
	clientID, ok := r.resourceClientMap[resourceID]
	if !ok {
		return nil, errUnmappedResource
	}

	c, ok := r.clients[clientID]
	if !ok {
		return nil, fmt.Errorf("failed to retrieve client [id=%s]: %w", clientID, errUnmappedClientID)
	}

	return c, nil
}

func (r *ResourceManager) getAllClientIDs() []string {
	return lo.Keys(r.clients)
}

//nolint:ireturn // returning interface to support generic type
func (r *ResourceManager) getClient(id string) (client.Client, error) {
	val, ok := r.clients[id]
	if !ok {
		return nil, fmt.Errorf("failed to get client [id=%s]: %w", id, errUnmappedClientID)
	}

	return val, nil
}

func (r *ResourceManager) fetchAllResourcesFromClient(clientID string) ([]string, error) {
	c, err := r.getClient(clientID)
	if err != nil {
		return nil, fmt.Errorf("failed to get client [id=%s]: %w", clientID, err)
	}

	// Fetch all volumes.
	log.Info().Msgf("[ClientId=%s] - Fetching volumes.", clientID)
	// TODO: better handle context.
	getVolResp, err := c.GetVolumes(context.Background(), &models.GetVolumesRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to get volumes from client [id=%s]: %w", clientID, err)
	}
	volumeIDs := lo.Map(getVolResp.Volumes, func(v *models.Volume, _ int) string {
		return v.UUID
	})
	log.Info().Msgf("[ClientId=%s] Fetched %d volumes.", clientID, len(volumeIDs))

	// Fetch all snapshots.
	log.Info().Msgf("[ClientId=%s] - Fetching snapshots.", clientID)
	// // TODO: better handle context
	getSnapshotResp, err := c.GetSnapshots(context.Background(), &models.GetSnapshotsRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshots from client [id=%s]: %w", clientID, err)
	}
	snapshotIDs := lo.Map(getSnapshotResp.Snapshots, func(s *models.Snapshot, _ int) string {
		return s.UUID
	})
	log.Info().Msgf("[ClientId=%s] Fetched %d snapshots.", clientID, len(snapshotIDs))

	return append(volumeIDs, snapshotIDs...), nil
}

func (r *ResourceManager) addResource(resourceID, clientID string) error {
	_, ok := r.resourceClientMap[resourceID]
	if ok {
		return errDuplicateResourceID
	}
	r.resourceClientMap[resourceID] = clientID

	return nil
}

func (r *ResourceManager) removeResource(resourceID string) error {
	_, ok := r.resourceClientMap[resourceID]
	if !ok {
		return errUnmappedClientID
	}
	delete(r.resourceClientMap, resourceID)

	return nil
}
