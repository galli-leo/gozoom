#include <wels/codec_def.h>
#include <wels/codec_api.h>
#include <wels/codec_app_def.h>
#include <string.h>
#include <stdio.h>

ISVCDecoder *pSvcDecoder = 0;

void DecodeOneFrame(unsigned char* input, int inputSize, unsigned char** output) {
    SBufferInfo sDstBufInfo;
    memset(&sDstBufInfo, 0, sizeof(SBufferInfo));
    if (pSvcDecoder == 0) {
        WelsCreateDecoder(&pSvcDecoder);
        SDecodingParam sDecParam = {0};
        sDecParam.sVideoProperty.eVideoBsType = VIDEO_BITSTREAM_AVC;
        //for Parsing only, the assignment is mandatory
        sDecParam.bParseOnly = false;
        (*pSvcDecoder)->Initialize(pSvcDecoder, &sDecParam);
        int option = WELS_LOG_DETAIL;
        (*pSvcDecoder)->SetOption(pSvcDecoder, DECODER_OPTION_TRACE_LEVEL, &option);
    }



    int iRet = (*pSvcDecoder)->DecodeFrameNoDelay(pSvcDecoder, input, inputSize, output, &sDstBufInfo);
    printf("return value: %d\n", iRet);
}