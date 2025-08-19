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

package commit

import (
	"context"
	"fmt"
	"time"

	"github.com/containerd/containerd/v2/core/diff"
	"github.com/containerd/containerd/v2/core/mount"
	"github.com/containerd/containerd/v2/core/snapshots"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

type clearCancel struct {
	context.Context
}

func (cc clearCancel) Deadline() (deadline time.Time, ok bool) {
	return
}

func (cc clearCancel) Done() <-chan struct{} {
	return nil
}

func (cc clearCancel) Err() error {
	return nil
}

// Background creates a new context which clears out the parent errors
func Background(ctx context.Context) context.Context {
	return clearCancel{ctx}
}

// Do runs the provided function with a context in which the
// errors are cleared out and will timeout after 10 seconds.
func Do(ctx context.Context, do func(context.Context)) {
	ctx, cancel := context.WithTimeout(clearCancel{ctx}, 10*time.Second)
	do(ctx)
	cancel()
}

// CreateDiff creates a layer diff for the given snapshot identifier from the
// parent of the snapshot. A content ref is provided to track the progress of
// the content creation and the provided snapshotter and mount differ are used
// for calculating the diff. The descriptor for the layer diff is returned.
func CreateDiff(ctx context.Context, snapshotID string, sn snapshots.Snapshotter, d diff.Comparer, removeTopLayer bool, opts ...diff.Opt) (ocispec.Descriptor, error) {
	info, err := sn.Stat(ctx, snapshotID)
	if err != nil {
		return ocispec.Descriptor{}, err
	}

	parent := info.Parent
	if removeTopLayer {
		secondInfo, err := sn.Stat(ctx, parent)
		if err != nil {
			return ocispec.Descriptor{}, err
		}
		if secondInfo.Parent != "" {
			parent = secondInfo.Parent
		}
	}

	lowerKey := fmt.Sprintf("%s-parent-view-%s", parent, uniquePart())
	lower, err := sn.View(ctx, lowerKey, parent)
	if err != nil {
		return ocispec.Descriptor{}, err
	}
	defer Do(ctx, func(ctx context.Context) {
		sn.Remove(ctx, lowerKey)
	})

	var upper []mount.Mount
	if info.Kind == snapshots.KindActive {
		upper, err = sn.Mounts(ctx, snapshotID)
		if err != nil {
			return ocispec.Descriptor{}, err
		}
	} else {
		upperKey := fmt.Sprintf("%s-view-%s", snapshotID, uniquePart())
		upper, err = sn.View(ctx, upperKey, snapshotID)
		if err != nil {
			return ocispec.Descriptor{}, err
		}
		defer Do(ctx, func(ctx context.Context) {
			sn.Remove(ctx, upperKey)
		})
	}

	return d.Compare(ctx, lower, upper, opts...)
}
