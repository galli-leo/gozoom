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
	"os"

	"github.com/galli-leo/gozoom/parser"
	"github.com/spf13/cobra"
)

func testH264(filename string) {
	f, err := os.Open(filename)
	if err != nil {
		logger.Fatalw("Failed to open file", "error", err, "filename", filename)
	}
	p := parser.NewH264Parser(f, logger)
	p.Parse()
	if p.Error() != nil {
		logger.Errorf("Had error during parsing: %v", p.Error())
	}
}

// h264Cmd represents the h264 command
var h264Cmd = &cobra.Command{
	Use:   "h264",
	Short: "Tests the h264 parser implementation",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		testH264(args[0])
	},
	Args: cobra.MaximumNArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"*.h264"}, cobra.ShellCompDirectiveDefault
	},
}

func init() {
	testCmd.AddCommand(h264Cmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// h264Cmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// h264Cmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
