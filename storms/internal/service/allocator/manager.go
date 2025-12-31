package allocator

import (
	"fmt"
	"math/rand/v2"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"gitlab.com/crusoeenergy/island/storage/storms/storms/internal/service/cluster"
)

type clusterManager interface {
	AllIDs() []string
	Get(string) (*cluster.Cluster, error)
}

type Manager struct {
	clusterManager clusterManager
}

func NewManager(cm clusterManager) *Manager {
	return &Manager{
		clusterManager: cm,
	}
}

func (a *Manager) AllocateCluster(affinityTags map[string]string) (string, error) {
	qualifiedClusters := []*cluster.Cluster{}
	clusterIDs := a.clusterManager.AllIDs()
	for _, clusterID := range clusterIDs {
		c, err := a.clusterManager.Get(clusterID)
		if err != nil {
			log.Warn().Err(err).Str("cluster_id", clusterID).Msg("could not consider cluster for allocation")

			continue
		}

		qualified := tagMatch(affinityTags, c.Config.AffinityTags)
		if qualified {
			qualifiedClusters = append(qualifiedClusters, c)
		}
	}

	if len(qualifiedClusters) == 0 {
		return "", fmt.Errorf("no qualified clusters")
	}

	if len(qualifiedClusters) == 1 {
		clusterID := qualifiedClusters[0].Config.ClusterID
		log.Info().Str("assigned_cluster_id", clusterID).Msg("identified 1 qualified cluster")

		return clusterID, nil
	}

	qualifiedClusterIDs := lo.Map[*cluster.Cluster, string](qualifiedClusters,
		func(c *cluster.Cluster, index int) string {
			return c.Config.ClusterID
		})

	// If there are more than one match, randomly assign!
	randomQualifiedClusterID := randomElement(qualifiedClusterIDs)

	clusterID := randomQualifiedClusterID
	log.Info().
		Str("assigned_cluster_id", clusterID).
		Strs("qualified_cluster_ids", qualifiedClusterIDs).
		Msgf("identified %d qualified cluster", len(qualifiedClusterIDs))

	return clusterID, nil
}

// Returns true if all key-value pairs in 'a' exist in 'b'.
func tagMatch(a, b map[string]string) bool {
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}

	return true
}

func randomElement[T any](list []T) T {
	return list[rand.IntN(len(list))]
}
