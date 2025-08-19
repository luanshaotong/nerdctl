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

// SnapshotListItem represents a snapshot in the list
type SnapshotListItem struct {
	Name        string         `json:"name"`
	Parent      string         `json:"parent,omitempty"`
	Kind        snapshots.Kind `json:"kind"`
	Created     time.Time      `json:"created"`
	Updated     time.Time      `json:"updated"`
	Snapshotter string         `json:"snapshotter"`
}

// List retrieves a list of snapshots
func List(ctx context.Context, client *containerd.Client, options types.SnapshotListOptions) error {
	var snapshotterName string
	if options.Snapshotter != "" {
		snapshotterName = options.Snapshotter
	} else {
		// Use default snapshotter if not specified
		snapshotterName = "overlayfs"
	}

	snapshotter := client.SnapshotService(snapshotterName)

	// Walk through all snapshots
	var snapshotList []SnapshotListItem
	err := snapshotter.Walk(ctx, func(ctx context.Context, info snapshots.Info) error {
		snapshotList = append(snapshotList, SnapshotListItem{
			Name:        info.Name,
			Parent:      info.Parent,
			Kind:        info.Kind,
			Created:     info.Created,
			Updated:     info.Updated,
			Snapshotter: snapshotterName,
		})
		return nil
	})

	if err != nil {
		log.L.WithError(err).Errorf("failed to walk snapshots")
		return fmt.Errorf("failed to list snapshots: %w", err)
	}

	// Output as JSON for now (can be extended to support different formats)
	encoder := json.NewEncoder(options.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(snapshotList)
}
