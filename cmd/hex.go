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
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"github.com/galli-leo/gozoom/zoom"
	"github.com/spf13/cobra"
)

var dumpLen uint = 0xa0
var dumpFile string = "dump.tex"
var dumpLineLen uint = 0x10
var dumpStart uint = 0x400

func newHexDumper(f io.Writer, sr *zoom.SampleReader) *hexDumper {
	return &hexDumper{f: f, sr: sr}
}

type hexDumper struct {
	f     io.Writer
	sr    *zoom.SampleReader
	total uint
	line  uint
}

func (h *hexDumper) writeLineStart() {
	h.f.Write([]byte(fmt.Sprintf("%04x & ", h.total+dumpStart)))
}

func (h *hexDumper) writeHexB(b byte, color string) {
	if h.total >= dumpLen {
		return
	}
	bStr := fmt.Sprintf("%02x", b)
	if color != "" {
		bStr = fmt.Sprintf(`\color{%s}{%s}`, color, bStr)
	}
	bStr = " " + bStr
	h.f.Write([]byte(bStr))
	h.total++
	h.line++
	if h.line >= dumpLineLen && h.total < dumpLen {
		h.f.Write([]byte(" \\\\\n"))
		h.writeLineStart()
		h.line = 0
	} else if h.total < dumpLen {
		h.f.Write([]byte(" & "))
	}
}

func (h *hexDumper) writeHex(bs []byte, color string) {
	for _, b := range bs {
		h.writeHexB(b, color)
	}
}

func (h *hexDumper) writeHexBin(data interface{}, color string) {
	buf := bytes.NewBuffer([]byte{})
	binary.Write(buf, binary.LittleEndian, data)
	logger.Infof("Buffer: %v", buf.Bytes())
	h.writeHex(buf.Bytes(), color)
}

var header uint32 = 0x2C05F158
var trailer uint32 = 0x84AD52E2

func (h *hexDumper) run() {
	h.writeLineStart()
	for h.sr.Next() {
		if h.total >= dumpLen {
			break
		}

		p := h.sr.Current()
		h.writeHexBin(header, "orange")
		h.writeHexBin(p.Type, "redflag")
		h.writeHex([]byte{0, 0, 0, 0}, "")
		h.writeHexBin(p.TimingA, "")
		h.writeHexBin(p.TimingB, "")
		h.writeHex([]byte{0, 0, 0, 0, 0, 0, 0, 0}, "")
		h.writeHexBin(p.DataSize, "brightgreen")
		h.writeHexBin(p.PropertySize, "blueedge")
		h.writeHex([]byte{0, 0, 0, 0, 0, 0, 0, 0}, "")
		h.writeHex(p.Property, "blueedge")
		h.writeHex(p.Data, "brightgreen")
		h.writeHexBin(trailer, "orange")
	}
}

func runHex(filename string) {
	logger.Infof("Dumping hex of file %s, length: %d, output: %s", filename, dumpLen, dumpFile)
	sr := zoom.NewSampleReader(logger)
	if err := sr.Open(filename); err != nil {
		logger.Fatalf("Failed to open file: %v", err)
	}
	outFile, err := os.OpenFile(dumpFile, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0777)
	defer outFile.Close()
	if err != nil {
		logger.Fatalf("Failed to open output file: %v", err)
	}
	h := newHexDumper(outFile, sr)
	h.run()
}

// hexCmd represents the hex command
var hexCmd = &cobra.Command{
	Use:   "hex",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		runHex(args[0])
	},
}

func init() {
	rootCmd.AddCommand(hexCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// hexCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// hexCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
