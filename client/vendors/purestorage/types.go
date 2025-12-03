package purestorage

type Reference struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type CreateVolumesResponse struct {
	Items []Volume `json:"items"`
}

type Volume struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Provisioned uint64     `json:"provisioned"`
	Created     uint64     `json:"created"`
	Serial      string     `json:"serial"`
	Source      *Reference `json:"source"`
}

type CreateSnapshotsResponse struct {
	Items []Snapshot `json:"items"`
}

type GetSnapshotsResponse struct {
	Items []Snapshot `json:"items"`
}

type Snapshot struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Provisioned uint64     `json:"provisioned"`
	Created     uint64     `json:"created"`
	Serial      string     `json:"serial"`
	Suffix      string     `json:"suffix"`
	Source      *Reference `json:"source"` // Reference to the source volume
}

type Host struct {
	Name string   `json:"name"`
	Nqns []string `json:"nqns"` // NVMe Qualified Names
}

type CreateHostsResponse struct {
	Items []Host `json:"items"`
}

type GetHostsResponse struct {
	Items []Host `json:"items"`
}

type Connection struct {
	Host   *Reference `json:"host"`
	Volume *Reference `json:"volume"`
	Lun    int        `json:"lun"`
}

type CreateConnectionsResponse struct {
	Items []Connection `json:"items"`
}

type GetConnectionsResponse struct {
	Items []Connection `json:"items"`
}
