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
	"fmt"
	"strings"

	containerd "github.com/containerd/containerd/v2/client"

	"github.com/containerd/nerdctl/v2/pkg/api/types"
)

// Update updates snapshot metadata (labels)
func Update(ctx context.Context, client *containerd.Client, snapshotID string, labelArgs []string, options types.SnapshotUpdateOptions) error {
	var snapshotterName string
	if options.Snapshotter != "" {
		snapshotterName = options.Snapshotter
	} else {
		// Use default snapshotter if not specified
		snapshotterName = "overlayfs"
	}

	snapshotter := client.SnapshotService(snapshotterName)

	// Parse label arguments (LABEL=VALUE format)
	labels := make(map[string]string)
	for _, labelArg := range labelArgs {
		parts := strings.SplitN(labelArg, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid label format %q (expected LABEL=VALUE)", labelArg)
		}
		labels[parts[0]] = parts[1]
	}

	// Get current snapshot info
	info, err := snapshotter.Stat(ctx, snapshotID)
	if err != nil {
		return fmt.Errorf("failed to get snapshot info: %w", err)
	}

	// Merge new labels with existing ones
	if info.Labels == nil {
		info.Labels = make(map[string]string)
	}
	for k, v := range labels {
		info.Labels[k] = v
	}

	// Update the snapshot
	_, err = snapshotter.Update(ctx, info, "labels")
	if err != nil {
		return fmt.Errorf("failed to update snapshot: %w", err)
	}

	fmt.Fprintf(options.Stdout, "Successfully updated snapshot %s\n", snapshotID)
	return nil
}
