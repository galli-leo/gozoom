/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

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
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// scan8Cmd represents the scan8 command
var scan8Cmd = &cobra.Command{
	Use:   "scan8",
	Short: "Display scan8 array",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		scan8 := [16*3 + 3]byte{
			4 + 1*8, 5 + 1*8, 4 + 2*8, 5 + 2*8,
			6 + 1*8, 7 + 1*8, 6 + 2*8, 7 + 2*8,
			4 + 3*8, 5 + 3*8, 4 + 4*8, 5 + 4*8,
			6 + 3*8, 7 + 3*8, 6 + 4*8, 7 + 4*8,
			4 + 6*8, 5 + 6*8, 4 + 7*8, 5 + 7*8,
			6 + 6*8, 7 + 6*8, 6 + 7*8, 7 + 7*8,
			4 + 8*8, 5 + 8*8, 4 + 9*8, 5 + 9*8,
			6 + 8*8, 7 + 8*8, 6 + 9*8, 7 + 9*8,
			4 + 11*8, 5 + 11*8, 4 + 12*8, 5 + 12*8,
			6 + 11*8, 7 + 11*8, 6 + 12*8, 7 + 12*8,
			4 + 13*8, 5 + 13*8, 4 + 14*8, 5 + 14*8,
			6 + 13*8, 7 + 13*8, 6 + 14*8, 7 + 14*8,
			0 + 0*8, 0 + 5*8, 0 + 10*8,
		}
		for i := 0; i < 16; i++ {
			fmt.Printf("%d ", scan8[i])
			if i%4 == 3 {
				fmt.Println()
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(scan8Cmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// scan8Cmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// scan8Cmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
