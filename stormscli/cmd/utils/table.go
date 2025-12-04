package utils

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
	admin "gitlab.com/crusoeenergy/island/storage/storms/pkg/api/gen/go/admin/v1"
	storms "gitlab.com/crusoeenergy/island/storage/storms/pkg/api/gen/go/storms/v1"
)

func RenderVolumes(volumes []*storms.Volume) error {
	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]string{"ID", "Size (bytes)", "Sector size", "ACL", "Available", "VendorVolumeID"})

	for _, v := range volumes {
		if err := table.Append([]string{
			v.Uuid,
			fmt.Sprintf("%d", v.Size),
			storms.SectorSizeEnum_name[int32(v.SectorSize)],
			strings.Join(v.Acl, ","),
			strconv.FormatBool(v.IsAvailable),
			v.GetVendorVolumeId(),
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
	table.Header([]string{"ID", "Size (bytes)", "Sector size", "SrcVolID", "Available", "VendorSnapshotID"})

	for _, v := range snapshots {
		if err := table.Append([]string{
			v.Uuid,
			fmt.Sprintf("%d bytes", v.Size),
			storms.SectorSizeEnum_name[int32(v.SectorSize)],
			v.SourceVolumeUuid,
			strconv.FormatBool(v.IsAvailable),
			v.GetVendorSnapshotId(),
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
