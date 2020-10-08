/*
Copyright © 2020 Leonardo Galli

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
	"os"
	"strings"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// func processCmdFile(path string) {
// 	sugar := newLogger("cmd_processor")
// 	filename := filepath.Join(path, "double_click_to_convert_02.zoom")
// 	f, err := OpenFile(filename)
// 	if err != nil {
// 		sugar.Fatalw("Failed to open file", "error", err, "filename", filename)
// 	}

// 	//f.ReadBeginning()
// 	p := NewCmdPacket()
// 	for f.HasData() {
// 		err := f.ReadPacket(p)
// 		if err != nil {
// 			sugar.Fatalw("Failed to read packet", "error", err)
// 		}
// 		sugar.Infof("Read packet: %v", p.CmdHeader)
// 		if p.AdditionalSize > 0 {
// 			sugar.Infof("Read packet: %s", p.AdditionalData)
// 		} else {
// 		}
// 	}
// }

// func processAudioFile(path string) {
// 	sugar := newLogger("audio_processor")
// 	filename := filepath.Join(path, "double_click_to_convert_01.zoom")
// 	f, err := OpenFile(filename)
// 	if err != nil {
// 		sugar.Fatalw("Failed to open file", "error", err, "filename", filename)
// 	}

// 	err = f.ReadBeginning()
// 	if err != nil {
// 		sugar.Fatalw("Failed to read beginning of file", "error", err)
// 	}
// 	p := NewSamplePacket()

// 	out, err := os.OpenFile("out.h264", os.O_CREATE|os.O_RDWR, 0777)
// 	if err != nil {
// 		sugar.Fatalw("Failed to open out file", "error", err)
// 	}

// 	timings := []int64{}
// 	positions := []CursorProp{}
// 	num := 0

// 	for f.HasData() {
// 		err := f.ReadPacket(p)
// 		if err != nil {
// 			sugar.Fatalw("Failed to read packet", "error", err)
// 		}
// 		if p.MediaType() == Audio {
// 			continue
// 		}
// 		sugar.Infof("Read packet: %v", p.SampleHeader)
// 		// if p.PropertySize > 0 {
// 		// 	sugar.Infof("Read property: %v", p.Property)
// 		// }

// 		var sps *parser.SPSInfo
// 		ppsInfo := make(map[uint]*parser.PPSInfo)
// 		spsInfo := make(map[uint]*parser.SPSInfo)
// 		if p.MediaType() == VideoScreenShare {
// 			sugar.Infof("Read video frame: %d", len(p.Data))
// 			_, err := out.Write(p.Data)
// 			if err != nil {
// 				sugar.Errorw("Failed to write packet to out file", "error", err)
// 			}
// 			_ = DecodeFrame(p.Data)
// 			f2, _ := os.OpenFile(fmt.Sprintf("frames/%d_%d.hex", num, 123), os.O_CREATE|os.O_RDWR, 0777)

// 			nalus, num := h264.SplitNALUs(p.Data)
// 			sugar.Infof("Split %d NALUS", num)
// 			for _, nalu := range nalus {
// 				t := h264.NALUType(nalu)
// 				sugar.Infof("Nalu type: %s", h264.NALUTypeString(t))
// 				if t == h264.NALU_SEI {
// 					sugar.Infof("SEI: %v", nalu)
// 					// out.Close()
// 					// os.Exit(0)
// 				}
// 				if t == h264.NALU_SPS {
// 					p := parser.NewSPSParser(nalu)
// 					sps = p.ParseInfo()
// 					spsInfo[sps.Id] = sps
// 				}
// 				if t == h264.NALU_PPS {
// 					p := parser.NewPPSParser(nalu, spsInfo)
// 					info := p.ParseInfo()
// 					ppsInfo[info.Id] = info
// 				}
// 				if t == h264.NALU_IDR {
// 					p := parser.NewSliceParser(nalu, sps, ppsInfo)
// 					slice := p.ParseHeader()
// 					_ = p.ParseSliceData()
// 					f2.Write(nalu)

// 					sugar.Infof("IDR: %+v", slice)
// 					sugar.Infof("DATA: %+v", p)
// 					os.Exit(0)
// 				}
// 			}
// 			f2.Close()
// 			sugar.Infof("video prop: %v", p.VideoProp())
// 		}
// 		timings = append(timings, p.TimingA)
// 		if p.MediaType() == Cursor {
// 			cur := p.CursorProp()
// 			positions = append(positions, *cur)
// 			img := image.NewRGBA(image.Rect(0, 0, int(cur.Width), int(cur.Height)))
// 			img.Pix = p.Data
// 			f, _ := os.OpenFile(fmt.Sprintf("images/%d_%d.png", num, cur.Type), os.O_RDWR|os.O_CREATE, 0777)
// 			png.Encode(f, img)
// 			f.Close()
// 			num++
// 		}
// 	}

// 	sugar.Infof("Got timings: %v", timings)
// 	sugar.Infof("Got Positions: %v", positions)

// 	out.Close()
// }

var logger *zap.SugaredLogger
var level string = zapcore.DebugLevel.String()
var levels = func() []string {
	ret := []string{}

	for start := zapcore.DebugLevel; start <= zapcore.FatalLevel; start++ {
		ret = append(ret, start.String())
	}

	return ret
}()

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gozoom",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		config := zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		if err := config.Level.UnmarshalText([]byte(level)); err != nil {
			return err
		}
		log, err := config.Build()
		logger = log.Sugar()
		return err
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		logger.Sync()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.gozoom.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	flags := rootCmd.PersistentFlags()
	validLevels := strings.Join(levels, ", ")
	flags.StringVarP(&level, "level", "l", level, fmt.Sprintf("Specifies the log level. If debug is selected, the logs will also include more information such as stack traces. Possible values: %s", validLevels))
}
