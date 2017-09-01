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
	"bufio"
	"fmt"
	"net"
	"os"

	"github.com/spf13/cobra"
)

// fakeclientCmd represents the fakeclient command
var fakeclientCmd = &cobra.Command{
	Use:   "fakeclient",
	Short: "a fake tcp client",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("fakeclient called")
		addr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:9090")
		conn, err := net.DialTCP("tcp", nil, addr)
		if err != nil {
			fmt.Println("err is %v", err.Error())
		}

		conn.SetNoDelay(true)

		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			conn.Write([]byte(fmt.Sprintf("%s\n", scanner.Text())))
		}
	},
}

func init() {
	RootCmd.AddCommand(fakeclientCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// fakeclientCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// fakeclientCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
