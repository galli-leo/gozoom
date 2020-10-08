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
	"strings"

	"github.com/galli-leo/gozoom/zoom"
	"github.com/spf13/cobra"
)

type ExtractType int

const (
	ExtractVideo ExtractType = iota
	ExtractAudio
)

//go:generate enumer -type=ExtractType -trimprefix=Extract

var extractType string = ExtractVideo.String()
var outputFile string = "out.h264"

func runExtract(filename string, method ExtractType) {
	logger.Infof("Extracting %s from %s to %s", method, filename, outputFile)
	sr := zoom.NewSampleReader(logger)
	if err := sr.Open(filename); err != nil {
		logger.Fatalf("Failed to open file: %v", err)
	}
	sd := zoom.NewSampleDataReader(sr, func(pkt *zoom.SamplePacket) bool {
		if method == ExtractAudio {
			if pkt.MediaType() == zoom.Audio {
				return true
			}
		}
		if pkt.MediaType() == zoom.VideoScreenShare {
			logger.Debugf("Timing: 0x%x, Name: 0x%x", pkt.TimingA, pkt.VideoProp().NameIdent)
			return true
		}
		if pkt.MediaType() != zoom.Audio {
			logger.Debugf("Media Type: %d", pkt.MediaType())
		}
		return false
	}, TransformFix, logger)
	outFile, err := os.OpenFile(outputFile, os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		logger.Fatalf("Failed to open output file: %v", err)
	}
	n, err := io.Copy(outFile, sd)
	logger.Infof("Written %d bytes to output file", n)
	if err != nil {
		logger.Errorf("Failed to write to output: %v", err)
	}
}

// extractCmd represents the extract command
var extractCmd = &cobra.Command{
	Use:   "extract",
	Short: "Extracts the specified data from the zoom file",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		t, err := ExtractTypeString(strings.Title(extractType))
		if err != nil {
			return err
		}
		runExtract(args[0], t)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(extractCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// extractCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// extractCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	extractCmd.Flags().StringVarP(&extractType, "type", "t", extractType, "Type of data to extract: video, audio")
	extractCmd.Flags().StringVarP(&outputFile, "out", "o", outputFile, "Output filename")
	extractCmd.MarkZshCompPositionalArgumentFile(1, "*.zoom")
}
