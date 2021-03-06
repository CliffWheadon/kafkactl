// Copyright © 2018 NAME HERE <jbonds@jbvm.io>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete Topics or Groups",
	Long: `Examples:
  kafkactl admin delete -t myTopic
  kafkactl admin delete -g myGroup`,
	Run: func(cmd *cobra.Command, args []string) {
		if cmd.Flags().Changed("topic") && cmd.Flags().Changed("group") {
			closeFatal("specify either group or topic, not both, try again.")
		}
		if cmd.Flags().Changed("topic") {
			deleteTopic(targetTopic)
			return
		}
		if cmd.Flags().Changed("group") {
			deleteGroup(targetGroup)
			return
		}
		closeFatal("specify either --group or --topic, try again.")
	},
}

func init() {
	adminCmd.AddCommand(deleteCmd)
}
