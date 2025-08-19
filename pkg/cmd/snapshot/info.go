/*
   Copyright The containerd Authors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package snapshot

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	containerd "github.com/containerd/containerd/v2/client"
	"github.com/containerd/containerd/v2/core/snapshots"
	"github.com/containerd/log"

	"github.com/containerd/nerdctl/v2/pkg/api/types"
)

// SnapshotInfo represents detailed information about a snapshot
type SnapshotInfo struct {
	Name        string            `json:"name"`
	Parent      string            `json:"parent,omitempty"`
	Kind        snapshots.Kind    `json:"kind"`
	Created     time.Time         `json:"created"`
	Updated     time.Time         `json:"updated"`
	Labels      map[string]string `json:"labels,omitempty"`
	Usage       *SnapshotUsage    `json:"usage,omitempty"`
	Snapshotter string            `json:"snapshotter"`
}

// SnapshotUsage represents usage information about a snapshot
type SnapshotUsage struct {
	Inodes int64 `json:"inodes"`
	Size   int64 `json:"size"`
}

// Info retrieves detailed information about a snapshot
func Info(ctx context.Context, client *containerd.Client, snapshotID string, options types.SnapshotInfoOptions) error {
	var snapshotterName string
	if options.Snapshotter != "" {
		snapshotterName = options.Snapshotter
	} else {
		// Use default snapshotter if not specified
		snapshotterName = "overlayfs"
	}

	snapshotter := client.SnapshotService(snapshotterName)

	// Get snapshot stat information
	info, err := snapshotter.Stat(ctx, snapshotID)
	if err != nil {
		return fmt.Errorf("failed to get snapshot info: %w", err)
	}

	// Get snapshot usage information
	usage, err := snapshotter.Usage(ctx, snapshotID)
	if err != nil {
		log.L.WithError(err).Warnf("failed to get snapshot usage for %s", snapshotID)
	}

	snapshotInfo := SnapshotInfo{
		Name:        info.Name,
		Parent:      info.Parent,
		Kind:        info.Kind,
		Created:     info.Created,
		Updated:     info.Updated,
		Labels:      info.Labels,
		Snapshotter: snapshotterName,
	}

	if err == nil {
		snapshotInfo.Usage = &SnapshotUsage{
			Inodes: usage.Inodes,
			Size:   usage.Size,
		}
	}

	// Output as JSON for now (can be extended to support different formats)
	encoder := json.NewEncoder(options.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(snapshotInfo)
}
