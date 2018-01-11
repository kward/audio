package wav

import (
	"fmt"
	"math"

	"github.com/kward/goaudio/codec/wav/encodings"
)

type CodecFormat struct {
	Encoding encodings.Encoding
	Bits     uint16
}

func (f CodecFormat) String() string {
	return fmt.Sprintf("{ encoding: %s bits: %d }", f.Encoding, f.Bits)
}

type Codec struct {
	ReadF32 func(src []byte, dst []float32, sampleCount int) (advance int)
}

var Codecs = map[CodecFormat]Codec{
	{encodings.PCM, 8}: {
		func(src []byte, dst []float32, sampleCount int) int {
			h := 0
			for k := 0; k < sampleCount; k++ {
				v := uint8(src[h])
				dst[k] = float32(v)/128.0 - 1.0
				h += 1
			}
			return h
		},
	},

	{encodings.PCM, 16}: {
		func(src []byte, dst []float32, sampleCount int) int {
			h := 0
			for k := 0; k < sampleCount; k++ {
				v := int16(src[h]) | int16(src[h+1])<<8
				dst[k] = float32(v) / float32(0x8000)
				h += 2
			}
			return h
		},
	},

	{encodings.PCM, 24}: {pcm24},

	{encodings.PCM, 32}: {
		func(src []byte, dst []float32, sampleCount int) int {
			h := 0
			for k := 0; k < sampleCount; k++ {
				v := int32(src[h]) | int32(src[h+1])<<8 | int32(src[h+2])<<16 | int32(src[h+3])<<24
				dst[k] = float32(v) / float32(0x80000000)
				h += 4
			}
			return h
		},
	},

	{encodings.EXTENSIBLE, 24}: {pcm24},

	{encodings.IEEE_Float, 32}: {
		func(src []byte, dst []float32, sampleCount int) int {
			h := 0
			for k := 0; k < sampleCount; k++ {
				v := uint32(src[h]) | uint32(src[h+1])<<8 | uint32(src[h+2])<<16 | uint32(src[h+3])<<24
				dst[k] = math.Float32frombits(v)
				h += 4
			}
			return h
		},
	},
}

func pcm24(src []byte, dst []float32, sampleCount int) int {
	h := 0
	for k := 0; k < sampleCount; k++ {
		v := int32(src[h]) | int32(src[h+1])<<8 | int32(src[h+2])<<16
		if v&0x800000 != 0 {
			v |= ^0xffffff
		}
		dst[k] = float32(v) / float32(0x800000)
		h += 3
	}
	return h
}
