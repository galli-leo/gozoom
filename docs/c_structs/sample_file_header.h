struct SampleFileHeader {
	uint32_t Header; // Always 0x2C05F158
	uint32_t Trailer; // Always 0x84AD52E2
    char unknown[24];
	uint32_t VersionInfo;
	uint32_t FileOffset;   // offset to start of actual data, always 1024
	char unused[56]; // Unknown and unused at the moment.
};