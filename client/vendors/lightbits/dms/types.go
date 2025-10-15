package dms

import "time"

// ClusterClusterStatus represents the connection status of the cluster.
type ClusterClusterStatus string

const (
	// ClusterStatusConnected means the cluster is connected.
	ClusterStatusConnected ClusterClusterStatus = "Connected"
	// ClusterStatusDisconnected means the cluster is disconnected.
	ClusterStatusDisconnected ClusterClusterStatus = "Disconnected"
	// ClusterStatusUnauthorized means the cluster connection is unauthorized.
	ClusterStatusUnauthorized ClusterClusterStatus = "Unauthorized"
)

// AttachClusterRequest is the request to attach a cluster to the DMS service.
// Title: Attach Cluster Request.
type AttachClusterRequest struct {
	// APIEndpoint is the management API endpoint of the cluster to connect to.
	// Any one of the API endpoints can be used and must be of the form `<host>:<port>`,
	// e.g.: `10.0.0.1:443`, `foo.bar.com:443`.
	APIEndpoint string `json:"apiEndpoint,omitempty"`
}

// AttachClusterResponse is the response to attach a Lightbits cluster to the DMS service.
// ListWorkflows API and the returned workflow ID can be used to track the status
// of this operation.
// Title: Attach Cluster Response.
type AttachClusterResponse struct {
	WorkflowID string `json:"workflowId,omitempty"`
}

// CancelWorkflowResponse is the response of cancel a workflow by its ID.
// Title: Cancel Workflow Response.
type CancelWorkflowResponse struct{}

// Cluster contains information about a Lightbits cluster connected to the DMS service.
type Cluster struct {
	// ID is the UUID of the cluster. Must be a valid UUID.
	ID string `json:"id,omitempty"`
	// Name is the name of the cluster.
	Name string `json:"name,omitempty"`
	// Version is the minimum version of the Lightbits release on the cluster.
	Version string `json:"version,omitempty"`
	// SubsystemNQN is the NQN of the subsystem of the cluster.
	SubsystemNQN string `json:"subsystemNqn,omitempty"`
	// APIEndpoints is the list of API endpoints of the cluster.
	APIEndpoints []string `json:"apiEndpoints,omitempty"`
	// DiscoveryEndpoints is the list of discovery endpoints of the cluster.
	DiscoveryEndpoints []string `json:"discoveryEndpoints,omitempty"`
	// ClusterConnectionStatus is the connection status of the cluster.
	ClusterConnectionStatus *ClusterConnectionStatus `json:"clusterConnectionStatus,omitempty"`
}

// ClusterConnectionStatus represents the connection status of the cluster.
type ClusterConnectionStatus struct {
	AccessConnectionStatus ClusterClusterStatus `json:"accessConnectionStatus,omitempty"`
	DataConnectionStatus   ClusterClusterStatus `json:"dataConnectionStatus,omitempty"`
}

// DetachClusterResponse is the response to detach a Lightbits cluster from the DMS service.
type DetachClusterResponse struct{}

// DstSnapshotInfo is the destination snapshot information for a thick clone snapshot operation.
// Title: The source snapshot identification information.
type DstSnapshotInfo struct {
	// Name is the name to assign to the destination (cloned) snapshot. It must
	// conform to valid Lightbits resource naming conventions up to 256
	// alpha-numeric characters, and must be either a letter, digit, hyphen,
	// underscore or period.
	Name string `json:"name,omitempty"`
	// ProjectName is the optional name of the project destination that the snapshot
	// will reside on. It must conform to a valid Lightbits resource naming conventions
	// up to 256 alpha numeric-characters, and must be either a letter, digit,
	// hyphen, underscore or period. If not specified, the destination snapshot will be
	// created in the same project as the source snapshot.
	ProjectName string `json:"projectName,omitempty"`
	// ClusterID is the UUID of the cluster where the destination snapshot will be
	// created (must be a valid UUID).
	ClusterID string `json:"clusterId,omitempty"`
	// Size is the optional size of the destination snapshot in bytes. If the size is
	// not specified, the size of the source snapshot's volume will be used.
	// If specified, the size must be equal or larger than the source snapshot's volume size.
	Size string `json:"size,omitempty"`
	// Replica is the optional number of replicas to create for the destination snapshot.
	// If not specified, the value will be taken from the source snapshot's volume.
	// If specified, the value must be between 1 and 3.
	Replica int32 `json:"replica,omitempty"`
	// Description is the optional description of the destination snapshot
	// (limited to up to 256 characters).
	Description string `json:"description,omitempty"`
	// QOSPolicyName is the optional name of an existing QoS policy to assign to the
	// destination snapshot. The QoS policy can be used to limit write late to the
	// destination volume.
	QOSPolicyName string `json:"qosPolicyName,omitempty"`
	// SectorSize is the optional sector size to create for the destination snapshot.
	// If not specified, the value will be taken from the source snapshot's volume.
	// If specified, the value must be either 512 or 4096.
	SectorSize int64 `json:"sectorSize,omitempty"`
	// Compression is an optional flag that determines whether compression is enabled
	// for the destination snapshot. If set to true, data stored in the volume will
	// be compressed. If not specified, compression settings will be inherited
	// from the source snapshot.
	Compression bool `json:"compression,omitempty"`
}

// DstVolumeInfo contains destination volume info for a thick clone operation.
type DstVolumeInfo struct {
	// Name is the name to assign to the destination (cloned) volume. It must
	// conform to valid Lightbits resource naming conventions up to 256 alpha numeric
	// characters, and must be either a letter, digit, hyphen, underscore or dot.
	Name string `json:"name,omitempty"`
	// ProjectName is the optional name of the project destination volume will reside on.
	// It must conform to valid Lightbits resource naming conventions up to 256
	// alpha-numeric characters, and must be either a letter, digit, hyphen,
	// underscore or a period. If not specified, the destination volume will be
	// created in the same project as the source snapshot.
	ProjectName string `json:"projectName,omitempty"`
	// ClusterID is the UUID of the cluster where the destination volume will be created
	// (must be a valid UUID).
	ClusterID string `json:"clusterId,omitempty"`
	// Size is the optional size of the destination volume in bytes. If the size is not
	// specified, the size of the source snapshot's volume will be used. If specified,
	// the size must be equal or larger than the source snapshot's volume size.
	Size string `json:"size,omitempty"`
	// Replica is the optional number of replicas to create for the destination volume.
	// If not specified the value will be taken from the source snapshot's volume.
	// If specified, the value must be between 1 and 3.
	Replica int32 `json:"replica,omitempty"`
	// QOSPolicyName is the optional name of an existing QoS policy to assign to the
	// destination volume. The QoS policy can be used to limit write late to the
	// destination volume.
	QOSPolicyName string `json:"qosPolicyName,omitempty"`
	// SectorSize is the optional sector size to create for the destination volume.
	// If not specified, the value will be taken from the source snapshot's volume.
	// If specified, the value must be either 512 or 4096.
	SectorSize int64 `json:"sectorSize,omitempty"`
	// Compression is an optional flag that determines whether compression is enabled
	// for the destination volume. If set to true, data stored in the volume will be
	// compressed. If not specified, compression settings will be inherited from
	// the source snapshot.
	Compression bool `json:"compression,omitempty"`
}

// GetServiceCredentialsResponse is the response for getting the service credentials
// for the DMS service.
// Title: Get Service Credentials Response.
type GetServiceCredentialsResponse struct {
	// PubKeyID is the key ID to be used while importing the public key into the
	// Lightbits cluster.
	PubKeyID string `json:"pubKeyId,omitempty"`
	// PubKey is the base64-encoded public key to be imported into the Lightbits cluster.
	//
	// After Base64 decoding, the contents should be saved as a file (e.g. into <FILE_PATH>)
	// and passed as an argument to the following command to the corresponding Lightbits
	// cluster, with <FILE_PATH> and <PUB_KEY_ID> substituted as appropriate.
	//   lbcli create credential --project-name=system --id=<PUB_KEY_ID> --type=rsa256pubkey <FILE_PATH>
	PubKey string `json:"pubKey,omitempty"`
}

// GetWorkflowResponse is the response to get a workflow by its ID.
// Title: Get Workflow Response.
type GetWorkflowResponse struct {
	// Workflow is the workflow information.
	Workflow *Workflow `json:"workflow,omitempty"`
}

// ListClustersResponse is the response to list the Lightbits clusters connected
// to the DMS service.
// Title: List Clusters Response.
type ListClustersResponse struct {
	Clusters []*Cluster `json:"clusters,omitempty"`
}

// ListWorkflowsResponse is the response to list the workflows. Note that the
// initial DMS service will only return the last 100 workflows.
// Title: List Workflows Response.
type ListWorkflowsResponse struct {
	Workflows     []*Workflow `json:"workflows,omitempty"`
	NextPageToken string      `json:"nextPageToken,omitempty"`
}

// LoginResponse is the response to the login to the DMS service.
// Title: Login Response.
type LoginResponse struct {
	// TokenType is the type of token.
	TokenType string `json:"tokenType,omitempty"`
	// ExpiresIn is the time in seconds when the token will expire.
	ExpiresIn string `json:"expiresIn,omitempty"`
	// IDToken is the ID of the token.
	IDToken string `json:"idToken,omitempty"`
}

// Progress is the progress information of a workflow.
// Title: Progress information.
type Progress struct {
	// Percent is the percentage of the workflow that has been completed.
	Percent int32 `json:"percent,omitempty"`
	// Stage is the current stage of the workflow.
	Stage string `json:"stage,omitempty"`
}

// ProtobufAny contains an arbitrary serialized message along with a URL that
// describes the type of the serialized message.
type ProtobufAny struct {
	Type                 string                 `json:"@type,omitempty"`
	AdditionalProperties map[string]interface{} `json:"-"` // Captured by unmarshaler
}

// RefreshClustersResponse is the response for refreshing Lightbits clusters information,
// returning an ID of the refresh workflow.
// Title: Refresh Clusters Response.
type RefreshClustersResponse struct {
	WorkflowID string `json:"workflowId,omitempty"`
}

// RPCStatus is the response for an unexpected error.
type RPCStatus struct {
	Code    int32          `json:"code,omitempty"`
	Message string         `json:"message,omitempty"`
	Details []*ProtobufAny `json:"details,omitempty"`
}

// SrcSnapshotInfo is the source snapshot identification information for a
// thick clone operation.
type SrcSnapshotInfo struct {
	// SnapID is the UUID of the snapshot to be cloned (must be a valid UUID).
	SnapID string `json:"snapId,omitempty"`
	// ProjectName is the name of the project source that the snapshot belongs to.
	// It must conform to valid Lightbits resource naming conventions up to 256
	// alpha-numeric characters, and must be either a letter, digit, hyphen,
	// underscore, or period.
	ProjectName string `json:"projectName,omitempty"`
	// ClusterID is the UUID of the cluster where the source snapshot is located
	// (must be a valid UUID).
	ClusterID string `json:"clusterId,omitempty"`
}

// ThickCloneSnapshotRequest is the request to create a thick clone snapshot from a source snapshot.
type ThickCloneSnapshotRequest struct {
	// Src is the source snapshot identification information for a thick clone
	// snapshot operation.
	Src *SrcSnapshotInfo `json:"src,omitempty"`
	// Dst is the destination (cloned) snapshot information for a thick clone
	// snapshot operation.
	Dst *DstSnapshotInfo `json:"dst,omitempty"`
	// VerifyClone is an optional flag. If set to true, the clone operation will
	// be verified by reading the destination volume.
	VerifyClone bool `json:"verifyClone,omitempty"`
}

// ThickCloneSnapshotResponse is the response to create a thick clone snapshot from a source snapshot.
// ListWorkflows API and the returned workflow ID can be used to track the status
// of this operation.
// Title: Thick Clone Snapshot Response.
type ThickCloneSnapshotResponse struct {
	WorkflowID string `json:"workflowId,omitempty"`
}

// ThickCloneVolumeRequest is the request to create a thick clone volume from a source snapshot.
// Title: Thick Clone Volume Request.
type ThickCloneVolumeRequest struct {
	// Src is the source snapshot identification information for a thick clone
	// volume operation.
	Src *SrcSnapshotInfo `json:"src,omitempty"`
	// Dst is the destination (cloned) volume information for a thick clone
	// volume operation.
	Dst *DstVolumeInfo `json:"dst,omitempty"`
	// VerifyClone is an optional flag. If set to true, the clone operation will
	// be verified by reading the destination volume.
	VerifyClone bool `json:"verifyClone,omitempty"`
}

// ThickCloneVolumeResponse is the response to create a thick clone volume from a source snapshot.
// ListWorkflows API and the returned workflow ID can be used to track the status
// of this operation.
// Title: Thick Clone Volume Response.
type ThickCloneVolumeResponse struct {
	WorkflowID string `json:"workflowId,omitempty"`
}

// Workflow holds information about a workflow.
type Workflow struct {
	// ID is the UUID of the workflow.
	ID string `json:"id,omitempty"`
	// CreatedAt is the time the workflow was created.
	CreatedAt time.Time `json:"createdAt,omitempty"`
	// StartedAt is the time the workflow started.
	StartedAt time.Time `json:"startedAt,omitempty"`
	// EndedAt is the time the workflow completed.
	EndedAt time.Time `json:"endedAt,omitempty"`
	// Type is the type of workflow (i.e., attach cluster, thick clone volume,
	// thick clone snapshot).
	Type string `json:"type,omitempty"`
	// State is the state of the workflow.
	State string `json:"state,omitempty"`
	// Msg is a detailed message providing additional information on the workflow
	// state. This is mainly used in case of failure.
	Msg string `json:"msg,omitempty"`
	// Progress is progress information of the workflow.
	Progress *Progress `json:"progress,omitempty"`
	// ThickCloneVolumeRequest holds the original request for a volume clone workflow.
	ThickCloneVolumeRequest *ThickCloneVolumeRequest `json:"thickCloneVolumeRequest,omitempty"`
	// SnapshotCloneRequest holds the original request for a snapshot clone workflow.
	SnapshotCloneRequest *ThickCloneSnapshotRequest `json:"snapshotCloneRequest,omitempty"`
	// AttachClusterRequest holds the original request for an attach cluster workflow.
	AttachClusterRequest *AttachClusterRequest `json:"attachClusterRequest,omitempty"`
}
