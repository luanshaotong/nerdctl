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
	"github.com/spf13/cobra"

	"github.com/containerd/nerdctl/v2/cmd/nerdctl/completion"
	"github.com/containerd/nerdctl/v2/cmd/nerdctl/helpers"
	"github.com/containerd/nerdctl/v2/pkg/api/types"
	"github.com/containerd/nerdctl/v2/pkg/clientutil"
	"github.com/containerd/nerdctl/v2/pkg/cmd/snapshot"
)

func UpdateCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:               "update [flags] SNAPSHOT_ID [LABEL=VALUE...]",
		Short:             "Update snapshot metadata",
		Args:              cobra.MinimumNArgs(1),
		RunE:              updateAction,
		ValidArgsFunction: updateShellComplete,
		SilenceUsage:      true,
		SilenceErrors:     true,
	}
	cmd.Flags().StringP("snapshotter", "", "", "Snapshotter name (default: auto-detect)")
	cmd.RegisterFlagCompletionFunc("snapshotter", completion.SnapshotterNames)
	return cmd
}

func updateOptions(cmd *cobra.Command) (types.SnapshotUpdateOptions, error) {
	globalOptions, err := helpers.ProcessRootCmdFlags(cmd)
	if err != nil {
		return types.SnapshotUpdateOptions{}, err
	}

	snapshotter, err := cmd.Flags().GetString("snapshotter")
	if err != nil {
		return types.SnapshotUpdateOptions{}, err
	}

	return types.SnapshotUpdateOptions{
		Stdout:      cmd.OutOrStdout(),
		GOptions:    globalOptions,
		Snapshotter: snapshotter,
	}, nil
}

func updateAction(cmd *cobra.Command, args []string) error {
	options, err := updateOptions(cmd)
	if err != nil {
		return err
	}
	client, ctx, cancel, err := clientutil.NewClient(cmd.Context(), options.GOptions.Namespace, options.GOptions.Address)
	if err != nil {
		return err
	}
	defer cancel()

	snapshotID := args[0]
	labels := args[1:]
	return snapshot.Update(ctx, client, snapshotID, labels, options)
}

func updateShellComplete(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// TODO: We could potentially add snapshot ID completion here in the future
	return nil, cobra.ShellCompDirectiveNoFileComp
}
