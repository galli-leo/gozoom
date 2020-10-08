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
	"io"
	"os"

	"github.com/galli-leo/gozoom/zoom"
	"github.com/spf13/cobra"
)

func runFile(filename string) {
	logger.Infof("testing file %s", filename)
	sr := zoom.NewSampleReader(logger)
	if err := sr.Open(filename); err != nil {
		logger.Fatalf("Failed to open file: %v", err)
	}
	for sr.Next() {
		curr := sr.Current()
		if curr.MediaType() == zoom.Avatar {
			logger.Infof("Have avatar packet: %+v", curr.Property)
			outf, err := os.OpenFile("avatar.png", os.O_CREATE|os.O_RDWR, 0777)
			if err != nil {
				logger.Fatalf("Failed to open file: %v", err)
			}
			defer outf.Close()
			_, err = outf.Write(curr.Data)
			if err != io.EOF && err != nil {
				logger.Fatalf("Failed to write data: %v", err)
			}
		}
	}
}

// fileCmd represents the file command
var fileCmd = &cobra.Command{
	Use:   "file",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		runFile(args[0])
	},
}

func init() {
	testCmd.AddCommand(fileCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// fileCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// fileCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
