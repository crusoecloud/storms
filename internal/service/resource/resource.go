package resource

type Type string

const (
	TypeVolume   Type = "volume"
	TypeSnapshot Type = "snapshot"
)

type Resource struct {
	ID           string
	ClusterID    string
	ResourceType Type
}
