package client

type Client interface {
	GetVolume([]string) error
	// CreateVolume([]string) error
	// ResizeVolume([]string) error
	// DeleteVolume([]string) error
	// AttachVolume([]string) error
	// DetachVolume([]string) error
	// GetSnapshot([]string) error
}
