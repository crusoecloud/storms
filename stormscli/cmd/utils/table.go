package utils

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
	admin "gitlab.com/crusoeenergy/island/storage/storms/pkg/api/gen/go/admin/v1"
	storms "gitlab.com/crusoeenergy/island/storage/storms/pkg/api/gen/go/storms/v1"
)

const (
	RFC3339 = "2006-01-02T15:04:05Z07:00"
)

func RenderVolumes(volumes []*storms.Volume) error {
	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]string{"ID", "Size (bytes)", "Sector size", "ACL", "Available", "VendorVolumeID", "CreatedAt"})

	for _, v := range volumes {
		createdAt := "N/A"
		if v.CreatedAt != nil {
			createdAt = v.CreatedAt.AsTime().Format(time.RFC3339)
		}
		if err := table.Append([]string{
			v.Uuid,
			formatBytes(v.Size),
			formatSectorSize(v.SectorSize),
			strings.Join(v.Acl, ","),
			strconv.FormatBool(v.IsAvailable),
			v.GetVendorVolumeId(),
			createdAt,
		}); err != nil {
			return fmt.Errorf("failed to append volume entry to table: %w", err)
		}
	}

	if err := table.Render(); err != nil {
		return fmt.Errorf("failed to render table: %w", err)
	}

	return nil
}

func RenderSnapshots(snapshots []*storms.Snapshot) error {
	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]string{"ID", "Size (bytes)", "Sector size", "SrcVolID", "Available", "VendorSnapshotID", "CreatedAt"})

	for _, v := range snapshots {
		createdAt := "N/A"
		if v.CreatedAt != nil {
			createdAt = v.CreatedAt.AsTime().Format(time.RFC3339)
		}
		if err := table.Append([]string{
			v.Uuid,
			fmt.Sprintf("%d bytes", v.Size),
			storms.SectorSizeEnum_name[int32(v.SectorSize)],
			v.SourceVolumeUuid,
			strconv.FormatBool(v.IsAvailable),
			v.GetVendorSnapshotId(),
			createdAt,
		}); err != nil {
			return fmt.Errorf("failed to append snapshot entry to table: %w", err)
		}
	}

	if err := table.Render(); err != nil {
		return fmt.Errorf("failed to render table: %w", err)
	}

	return nil
}

func RenderClusters(clusters []*admin.Cluster) error {
	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]string{"ClusterID", "Vendor", "Num Volumes", "Num Snapshots"})

	for _, cluster := range clusters {
		if err := table.Append([]string{
			cluster.Id,
			cluster.Vendor,
			strconv.FormatInt(int64(cluster.ResourceCount["volume"]), 10),   // TODO - vheng import this from somewhere..
			strconv.FormatInt(int64(cluster.ResourceCount["snapshot"]), 10), // TODO - vheng import this from some where
		}); err != nil {
			return fmt.Errorf("failed to append cluster entry to table: %w", err)
		}
	}

	if err := table.Render(); err != nil {
		return fmt.Errorf("failed to render table: %w", err)
	}

	return nil
}

func formatBytes(b uint64) string {
	// EiB is 2^60, and we cannot go any higher with uint64
	units := []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB", "EiB"}

	// Handle 0 explicitly
	if b == 0 {
		return "0 B"
	}

	size := float64(b)
	i := 0

	// While size > 1024 and we haven't run out of units
	for size >= 1024 && i < len(units)-1 {
		size /= 1024
		i++
	}

	return fmt.Sprintf("%.2f %s", size, units[i])
}

func formatSectorSize(enum storms.SectorSizeEnum) string {
	//nolint:exhaustive // default case handles unspecified
	switch enum {
	case storms.SectorSizeEnum_SECTOR_SIZE_ENUM_4096:
		return "4096 B"
	case storms.SectorSizeEnum_SECTOR_SIZE_ENUM_512:
		return "512 B"
	default:
		return "unspecified"
	}
}
