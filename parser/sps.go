package parser

import (
	"bytes"
	"math"

	"github.com/nareix/joy5/codec/h264"
	"github.com/nareix/joy5/utils/bits"
	"go.uber.org/zap"
)

func NewSPSParser(p *H264Parser) *SPSParser {
	return &SPSParser{
		p.GolombBitReader,
		p.log.Named("SPSParser"),
		p,
	}
}

type SPSParser struct {
	*GolombBitReader
	log  *zap.SugaredLogger
	h264 *H264Parser
}

type SPSInfo struct {
	Id                uint
	ProfileIdc        uint
	LevelIdc          uint
	ConstraintSetFlag uint

	Log2MaxFrameNumMinus4       uint
	PicOrderCntType             uint
	Log2MaxPicOrderCntLsbMinus4 uint

	BitDepthLumaMinus8   uint
	BitDepthChromaMinus8 uint

	ChromaFormatIdc         uint
	SeparateColourPlaneFlag uint

	PicWidthInMbsMinus1       uint
	PicHeightInMapUnitsMinus1 uint
	FrameMbsOnlyFlag          uint
	MbAdaptiveFrameFieldFlag  uint
	// MbWidth                   uint
	// MbHeight                  uint

	CropLeft   uint
	CropRight  uint
	CropTop    uint
	CropBottom uint

	Width  uint
	Height uint

	FPS uint
}

func (s *SPSInfo) BitDepthY() uint {
	return s.BitDepthLumaMinus8 + 8
}

func (s *SPSInfo) BitDepthC() uint {
	return s.BitDepthChromaMinus8 + 8
}

func (s *SPSInfo) SubWidthC() uint {
	switch s.ChromaFormatIdc {
	case 1:
		return 2
	case 2:
		return 2
	case 3:
		return 1
	}
	return 0
}

func (s *SPSInfo) SubHeightC() uint {
	switch s.ChromaFormatIdc {
	case 1:
		return 2
	case 2:
		return 1
	case 3:
		return 1
	}
	return 0
}

func (s *SPSInfo) NumC8x8() uint {
	return 4 / (s.SubWidthC() * s.SubHeightC())
}

func (s *SPSInfo) MbWidthC() uint {
	if s.ChromaFormatIdc == 0 {
		return 0
	}

	return 16 / s.SubWidthC()
}

func (s *SPSInfo) MbHeightC() uint {
	if s.ChromaFormatIdc == 0 {
		return 0
	}

	return 16 / s.SubWidthC()
}

func (s *SPSInfo) PicWidthInMbs() uint {
	return s.PicWidthInMbsMinus1 + 1
}

func (s *SPSInfo) PicWidthInSamplesL() uint {
	return s.PicWidthInMbs() * 16
}

func (s *SPSInfo) PicHeightInMapUnits() uint {
	return s.PicHeightInMapUnitsMinus1 + 1
}

func (s *SPSInfo) PicSizeInMapUnits() uint {
	return s.PicWidthInMbs() * s.PicHeightInMapUnits()
}

func (s *SPSInfo) FrameHeightInMbs() uint {
	return (2 - s.FrameMbsOnlyFlag) * s.PicHeightInMapUnits()
}

func (s *SPSInfo) PicHeightInMbs() uint {
	return s.FrameHeightInMbs()
}

func (s *SPSInfo) PicSizeInMbs() uint {
	return s.PicWidthInMbs() * s.PicHeightInMbs()
}

func (s *SPSInfo) ChromaArrayType() uint {
	if s.SeparateColourPlaneFlag == 0 {
		return s.ChromaFormatIdc
	}
	return 0
}

func (p *SPSParser) ParseInfo() *SPSInfo {
	s := &SPSInfo{}

	s.ProfileIdc = p.ReadBits(8)
	// constraint_set0_flag-constraint_set6_flag,reserved_zero_2bits
	s.ConstraintSetFlag = p.ReadBits(8)
	s.ConstraintSetFlag = s.ConstraintSetFlag >> 2

	s.LevelIdc = p.ReadBits(8)

	// seq_parameter_set_id
	s.Id = p.ReadUE()

	if s.ProfileIdc == 100 || s.ProfileIdc == 110 ||
		s.ProfileIdc == 122 || s.ProfileIdc == 244 ||
		s.ProfileIdc == 44 || s.ProfileIdc == 83 ||
		s.ProfileIdc == 86 || s.ProfileIdc == 118 ||
		s.ProfileIdc == 128 || s.ProfileIdc == 138 ||
		s.ProfileIdc == 139 || s.ProfileIdc == 134 ||
		s.ProfileIdc == 135 {

		s.ChromaFormatIdc = p.ReadUE()

		if s.ChromaFormatIdc == 3 {
			// separate_colour_plane_flag
			s.SeparateColourPlaneFlag = p.ReadBits(1)
		}

		// bit_depth_luma_minus8
		s.BitDepthLumaMinus8 = p.ReadUE()
		// bit_depth_chroma_minus8
		s.BitDepthChromaMinus8 = p.ReadUE()
		// qpprime_y_zero_transform_bypass_flag
		p.ReadBits(1)

		var seq_scaling_matrix_present_flag uint
		seq_scaling_matrix_present_flag = p.ReadBits(1)

		if seq_scaling_matrix_present_flag != 0 {
			for i := 0; i < 8; i++ {
				var seq_scaling_list_present_flag uint
				seq_scaling_list_present_flag = p.ReadBits(1)
				if seq_scaling_list_present_flag != 0 {
					var sizeOfScalingList uint
					if i < 6 {
						sizeOfScalingList = 16
					} else {
						sizeOfScalingList = 64
					}
					lastScale := int(8)
					nextScale := int(8)
					for j := uint(0); j < sizeOfScalingList; j++ {
						if nextScale != 0 {
							var delta_scale int
							delta_scale = p.ReadSE()
							nextScale = (lastScale + delta_scale + 256) % 256
						}
						if nextScale != 0 {
							lastScale = nextScale
						}
					}
				}
			}
		}

	}

	// log2_max_frame_num_minus4
	s.Log2MaxFrameNumMinus4 = p.ReadUE()

	var pic_order_cnt_type uint
	pic_order_cnt_type = p.ReadUE()
	s.PicOrderCntType = pic_order_cnt_type
	if pic_order_cnt_type == 0 {
		// log2_max_pic_order_cnt_lsb_minus4
		s.Log2MaxPicOrderCntLsbMinus4 = p.ReadUE()
	} else if pic_order_cnt_type == 1 {
		// delta_pic_order_always_zero_flag
		p.ReadBits(1)
		// offset_for_non_ref_pic
		p.ReadSE()
		// offset_for_top_to_bottom_field
		p.ReadSE()
		var num_ref_frames_in_pic_order_cnt_cycle uint
		num_ref_frames_in_pic_order_cnt_cycle = p.ReadUE()
		for i := uint(0); i < num_ref_frames_in_pic_order_cnt_cycle; i++ {
			p.ReadSE()
		}
	}

	// max_num_ref_frames
	p.ReadUE()

	// gaps_in_frame_num_value_allowed_flag
	p.ReadBits(1)

	s.PicWidthInMbsMinus1 = p.ReadUE()
	//s.MbWidth++

	s.PicHeightInMapUnitsMinus1 = p.ReadUE()
	//s.MbHeight++

	s.FrameMbsOnlyFlag = p.ReadBits(1)
	if s.FrameMbsOnlyFlag == 0 {
		s.MbAdaptiveFrameFieldFlag = p.ReadBits(1)
	}

	// direct_8x8_inference_flag
	p.ReadBits(1)

	var frame_cropping_flag uint
	frame_cropping_flag = p.ReadBits(1)
	if frame_cropping_flag != 0 {
		s.CropLeft = p.ReadUE()
		s.CropRight = p.ReadUE()
		s.CropTop = p.ReadUE()
		s.CropBottom = p.ReadUE()
	}

	// s.Width = (s.MbWidth * 16) - s.CropLeft*2 - s.CropRight*2
	// s.Height = ((2 - frame_mbs_only_flag) * s.MbHeight * 16) - s.CropTop*2 - s.CropBottom*2

	vui_parameter_present_flag := p.ReadBits(1)

	if vui_parameter_present_flag != 0 {
		p.ParseVUIParameters()
	}

	return s
}

type VUIParameters struct {
}

func (p *SPSParser) ParseVUIParameters() *VUIParameters {
	v := &VUIParameters{}

	aspect_ratio_info_present_flag := p.ReadBits(1)

	if aspect_ratio_info_present_flag != 0 {
		aspect_ratio_idc := p.ReadBits(8)

		if aspect_ratio_idc == 255 {
			sar_width := p.ReadBits(16)
			sar_height := p.ReadBits(16)

			_, _ = sar_width, sar_height
		}
	}

	overscan_info_present_flag := p.ReadBits(1)

	if overscan_info_present_flag != 0 {
		overscan_appropriate_flagu := p.ReadBits(1)

		_ = overscan_appropriate_flagu
	}

	video_signal_type_present_flag := p.ReadBits(1)
	if video_signal_type_present_flag != 0 {
		video_format := p.ReadBits(3)
		_ = video_format
		video_full_range_flag := p.ReadBits(1)
		_ = video_full_range_flag
		colour_description_present_flag := p.ReadBits(1)
		if colour_description_present_flag != 0 {
			colour_primaries := p.ReadBits(8)
			_ = colour_primaries
			transfer_characteristics := p.ReadBits(8)
			_ = transfer_characteristics
			matrix_coefficients := p.ReadBits(8)
			_ = matrix_coefficients
		}
	}
	chroma_loc_info_present_flag := p.ReadBits(1)
	if chroma_loc_info_present_flag != 0 {
		chroma_sample_loc_type_top_field := p.ReadSE()
		_ = chroma_sample_loc_type_top_field

		chroma_sample_loc_type_bottom_field := p.ReadSE()
		_ = chroma_sample_loc_type_bottom_field
	}

	timing_info_present_flag := p.ReadBits(1)
	if timing_info_present_flag != 0 {
		num_units_in_tick := p.ReadBits(32)
		time_scale := p.ReadBits(32)

		//num_units_in_tick/time_scale should yield field time  要求向下取整
		fps := uint(math.Floor(float64(time_scale) / float64(num_units_in_tick) / 2.0))

		fixed_frame_rate_flag := p.ReadBits(1)
		if fixed_frame_rate_flag != 0 {
			//utils.L.InfoLn("fixed_frame_rate_flag", fixed_frame_rate_flag)
			//have been devide 2
			//self.FPS = self.FPS / 2
		}
		_ = fps
	}

	return v
}

func ParseSPS(data []byte) (s SPSInfo, err error) {
	data = h264.RemoveH264orH265EmulationBytes(data)
	r := &bits.GolombBitReader{R: bytes.NewReader(data)}

	if _, err = r.ReadBits(8); err != nil {
		return
	}

	if s.ProfileIdc, err = r.ReadBits(8); err != nil {
		return
	}

	// constraint_set0_flag-constraint_set6_flag,reserved_zero_2bits
	if s.ConstraintSetFlag, err = r.ReadBits(8); err != nil {
		return
	}
	s.ConstraintSetFlag = s.ConstraintSetFlag >> 2

	// level_idc
	if s.LevelIdc, err = r.ReadBits(8); err != nil {
		return
	}

	// seq_parameter_set_id
	if s.Id, err = r.ReadExponentialGolombCode(); err != nil {
		return
	}

	if s.ProfileIdc == 100 || s.ProfileIdc == 110 ||
		s.ProfileIdc == 122 || s.ProfileIdc == 244 ||
		s.ProfileIdc == 44 || s.ProfileIdc == 83 ||
		s.ProfileIdc == 86 || s.ProfileIdc == 118 {

		if s.ChromaFormatIdc, err = r.ReadExponentialGolombCode(); err != nil {
			return
		}

		if s.ChromaFormatIdc == 3 {
			// residual_colour_transform_flag
			if _, err = r.ReadBit(); err != nil {
				return
			}
		}

		// bit_depth_luma_minus8
		if s.BitDepthLumaMinus8, err = r.ReadExponentialGolombCode(); err != nil {
			return
		}
		// bit_depth_chroma_minus8
		if s.BitDepthChromaMinus8, err = r.ReadExponentialGolombCode(); err != nil {
			return
		}
		// qpprime_y_zero_transform_bypass_flag
		if _, err = r.ReadBit(); err != nil {
			return
		}

		var seq_scaling_matrix_present_flag uint
		if seq_scaling_matrix_present_flag, err = r.ReadBit(); err != nil {
			return
		}

		if seq_scaling_matrix_present_flag != 0 {
			for i := 0; i < 8; i++ {
				var seq_scaling_list_present_flag uint
				if seq_scaling_list_present_flag, err = r.ReadBit(); err != nil {
					return
				}
				if seq_scaling_list_present_flag != 0 {
					var sizeOfScalingList uint
					if i < 6 {
						sizeOfScalingList = 16
					} else {
						sizeOfScalingList = 64
					}
					lastScale := uint(8)
					nextScale := uint(8)
					for j := uint(0); j < sizeOfScalingList; j++ {
						if nextScale != 0 {
							var delta_scale uint
							if delta_scale, err = r.ReadSE(); err != nil {
								return
							}
							nextScale = (lastScale + delta_scale + 256) % 256
						}
						if nextScale != 0 {
							lastScale = nextScale
						}
					}
				}
			}
		}
	}

	// log2_max_frame_num_minus4
	if s.Log2MaxFrameNumMinus4, err = r.ReadExponentialGolombCode(); err != nil {
		return
	}

	var pic_order_cnt_type uint
	if pic_order_cnt_type, err = r.ReadExponentialGolombCode(); err != nil {
		return
	}
	s.PicOrderCntType = pic_order_cnt_type
	if pic_order_cnt_type == 0 {
		// log2_max_pic_order_cnt_lsb_minus4
		if s.Log2MaxPicOrderCntLsbMinus4, err = r.ReadExponentialGolombCode(); err != nil {
			return
		}
	} else if pic_order_cnt_type == 1 {
		// delta_pic_order_always_zero_flag
		if _, err = r.ReadBit(); err != nil {
			return
		}
		// offset_for_non_ref_pic
		if _, err = r.ReadSE(); err != nil {
			return
		}
		// offset_for_top_to_bottom_field
		if _, err = r.ReadSE(); err != nil {
			return
		}
		var num_ref_frames_in_pic_order_cnt_cycle uint
		if num_ref_frames_in_pic_order_cnt_cycle, err = r.ReadExponentialGolombCode(); err != nil {
			return
		}
		for i := uint(0); i < num_ref_frames_in_pic_order_cnt_cycle; i++ {
			if _, err = r.ReadSE(); err != nil {
				return
			}
		}
	}

	// max_num_ref_frames
	if _, err = r.ReadExponentialGolombCode(); err != nil {
		return
	}

	// gaps_in_frame_num_value_allowed_flag
	if _, err = r.ReadBit(); err != nil {
		return
	}

	// if s.MbWidth, err = r.ReadExponentialGolombCode(); err != nil {
	// 	return
	// }
	// s.MbWidth++

	// if s.MbHeight, err = r.ReadExponentialGolombCode(); err != nil {
	// 	return
	// }
	// s.MbHeight++

	var frame_mbs_only_flag uint
	if frame_mbs_only_flag, err = r.ReadBit(); err != nil {
		return
	}
	if frame_mbs_only_flag == 0 {
		// mb_adaptive_frame_field_flag
		if _, err = r.ReadBit(); err != nil {
			return
		}
	}

	// direct_8x8_inference_flag
	if _, err = r.ReadBit(); err != nil {
		return
	}

	var frame_cropping_flag uint
	if frame_cropping_flag, err = r.ReadBit(); err != nil {
		return
	}
	if frame_cropping_flag != 0 {
		if s.CropLeft, err = r.ReadExponentialGolombCode(); err != nil {
			return
		}
		if s.CropRight, err = r.ReadExponentialGolombCode(); err != nil {
			return
		}
		if s.CropTop, err = r.ReadExponentialGolombCode(); err != nil {
			return
		}
		if s.CropBottom, err = r.ReadExponentialGolombCode(); err != nil {
			return
		}
	}

	// s.Width = (s.MbWidth * 16) - s.CropLeft*2 - s.CropRight*2
	// s.Height = ((2 - frame_mbs_only_flag) * s.MbHeight * 16) - s.CropTop*2 - s.CropBottom*2

	vui_parameter_present_flag, err := r.ReadBit()
	if err != nil {
		return
	}

	if vui_parameter_present_flag != 0 {
		aspect_ratio_info_present_flag, err := r.ReadBit()
		if err != nil {
			return s, err
		}

		if aspect_ratio_info_present_flag != 0 {
			aspect_ratio_idc, err := r.ReadBits(8)
			if err != nil {
				return s, err
			}

			if aspect_ratio_idc == 255 {
				sar_width, err := r.ReadBits(16)
				if err != nil {
					return s, err
				}
				sar_height, err := r.ReadBits(16)
				if err != nil {
					return s, err
				}

				_, _ = sar_width, sar_height
			}
		}

		overscan_info_present_flag, err := r.ReadBit()
		if err != nil {
			return s, err
		}

		if overscan_info_present_flag != 0 {
			overscan_appropriate_flagu, err := r.ReadBit()
			if err != nil {
				return s, err
			}

			_ = overscan_appropriate_flagu
		}
		video_signal_type_present_flag, err := r.ReadBit()
		if video_signal_type_present_flag != 0 {
			video_format, err := r.ReadBits(3)
			if err != nil {
				return s, err
			}
			_ = video_format
			video_full_range_flag, err := r.ReadBit()
			if err != nil {
				return s, err
			}
			_ = video_full_range_flag
			colour_description_present_flag, err := r.ReadBit()
			if err != nil {
				return s, err
			}
			if colour_description_present_flag != 0 {
				colour_primaries, err := r.ReadBits(8)
				if err != nil {
					return s, err
				}
				_ = colour_primaries
				transfer_characteristics, err := r.ReadBits(8)
				if err != nil {
					return s, err
				}
				_ = transfer_characteristics
				matrix_coefficients, err := r.ReadBits(8)
				if err != nil {
					return s, err
				}
				_ = matrix_coefficients
			}
		}
		chroma_loc_info_present_flag, err := r.ReadBit()
		if err != nil {
			return s, err
		}
		if chroma_loc_info_present_flag != 0 {
			chroma_sample_loc_type_top_field, err := r.ReadSE()
			if err != nil {
				return s, err
			}
			_ = chroma_sample_loc_type_top_field

			chroma_sample_loc_type_bottom_field, err := r.ReadSE()
			if err != nil {
				return s, err
			}

			_ = chroma_sample_loc_type_bottom_field
		}

		timing_info_present_flag, err := r.ReadBit()
		if err != nil {
			return s, err
		}

		if timing_info_present_flag != 0 {
			num_units_in_tick, err := r.ReadBits(32)
			if err != nil {
				return s, err
			}
			time_scale, err := r.ReadBits(32)
			if err != nil {
				return s, err
			}

			//num_units_in_tick/time_scale should yield field time  要求向下取整
			s.FPS = uint(math.Floor(float64(time_scale) / float64(num_units_in_tick) / 2.0))

			fixed_frame_rate_flag, err := r.ReadBit()
			if err != nil {
				return s, err
			}
			if fixed_frame_rate_flag != 0 {
				//utils.L.InfoLn("fixed_frame_rate_flag", fixed_frame_rate_flag)
				//have been devide 2
				//self.FPS = self.FPS / 2
			}
		}
	}

	return
}
