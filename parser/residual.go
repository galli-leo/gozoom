package parser

const (
	NUM_LEVELS     = 16
	NUM_COEFFS     = 16
	NUM_8x8_LEVELS = 4
	NUM_8x8_COEFFS = 64
)

type ResidualBlock struct {
	cat    BlockCat
	blkIdx uint
	level  []int
}

type ResidualData struct {
	cat     BlockCat
	levelDC [NUM_COEFFS]int
	levelAC [NUM_LEVELS][NUM_COEFFS]int
	level4  [NUM_LEVELS][NUM_COEFFS]int
	level8  [NUM_8x8_LEVELS][NUM_8x8_COEFFS]int
}

func (r *ResidualData) GetBlock(cat BlockCat, idx uint) *ResidualBlock {
	c := r.cat | cat
	var level []int

	switch cat.Size() {
	case BlockDC:
		level = r.levelDC[:]
	case BlockAC:
		level = r.levelAC[idx][:]
	case BlockLevel:
		level = r.level4[idx][:]
	case BlockLevel8:
		level = r.level8[idx][:]
	}

	return &ResidualBlock{
		c,
		idx,
		level,
	}
}
