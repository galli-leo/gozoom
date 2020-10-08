struct CmdHeader {
	uint32_t SizeToRead;
	int32_t Type;
    char unknown[8];
	int32_t NameIdent;
    char padding[4];
	uint32_t TimingA;
    char unknown[22];
	uint16_t SomeType;
    char padding[4];
	int32_t AdditionalSize;
    char padding[4];
};