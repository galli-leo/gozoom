struct SampleHeader {
	int32_t Type;
    char padding[4];
	int64_t TimingA;
	int64_t TimingB;
    char unknown[8];
    int32_t DataSize;
    int32_t PropertySize;
    char unknown[8];
};