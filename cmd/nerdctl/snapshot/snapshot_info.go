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

func InfoCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:               "info [flags] SNAPSHOT_ID",
		Short:             "Get detailed information about a snapshot",
		Args:              helpers.IsExactArgs(1),
		RunE:              infoAction,
		ValidArgsFunction: infoShellComplete,
		SilenceUsage:      true,
		SilenceErrors:     true,
	}
	cmd.Flags().StringP("snapshotter", "", "", "Snapshotter name (default: auto-detect)")
	cmd.RegisterFlagCompletionFunc("snapshotter", completion.SnapshotterNames)
	return cmd
}

func infoOptions(cmd *cobra.Command) (types.SnapshotInfoOptions, error) {
	globalOptions, err := helpers.ProcessRootCmdFlags(cmd)
	if err != nil {
		return types.SnapshotInfoOptions{}, err
	}

	snapshotter, err := cmd.Flags().GetString("snapshotter")
	if err != nil {
		return types.SnapshotInfoOptions{}, err
	}

	return types.SnapshotInfoOptions{
		Stdout:      cmd.OutOrStdout(),
		GOptions:    globalOptions,
		Snapshotter: snapshotter,
	}, nil
}

func infoAction(cmd *cobra.Command, args []string) error {
	options, err := infoOptions(cmd)
	if err != nil {
		return err
	}
	client, ctx, cancel, err := clientutil.NewClient(cmd.Context(), options.GOptions.Namespace, options.GOptions.Address)
	if err != nil {
		return err
	}
	defer cancel()

	return snapshot.Info(ctx, client, args[0], options)
}

func infoShellComplete(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// TODO: We could potentially add snapshot ID completion here in the future
	return nil, cobra.ShellCompDirectiveNoFileComp
}
