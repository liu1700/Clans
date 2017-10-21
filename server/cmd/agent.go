// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
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
	"Clans/server/agent"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

// agentCmd represents the agent command
var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "代理服务器",
	Long:  `代理服务器，唯一面向外网的服务器`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("agent called")

		config := &agent.Config{
			Listen:       "192.168.1.102:9090",
			ReadDeadline: 15 * time.Second,
			Sockbuf:      32767,
			Udp_sockbuf:  4194304,
			Txqueuelen:   128,
			Dscp:         46,
			Sndwnd:       32,
			Rcvwnd:       32,
			Mtu:          1280,
			Nodelay:      1,
			Interval:     20,
			Resend:       1,
			Nc:           1,
			RpmLimit:     200,
		}

		agent.Start(config)
	},
}

func init() {
	RootCmd.AddCommand(agentCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// agentCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// agentCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
