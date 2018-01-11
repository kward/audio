/*
https://en.wikipedia.org/wiki/Resource_Interchange_File_Format
https://en.wikipedia.org/wiki/WAV
http://soundfile.sapp.org/doc/WaveFormat/
*/
package wav

import (
	"github.com/kward/goaudio/codec/wav/encodings"
)

func findSubChunk(data []byte, target [4]byte) []byte {
	var id [4]byte
	var size uint32
	for len(data) >= 8 {
		copy(id[:], data)
		readU32LE(&size, data[4:])
		if id == target {
			return data
		}

		chunkSize := int(size)
		if id == [4]byte{'R', 'I', 'F', 'F'} {
			chunkSize = 4
		}

		data = data[8+chunkSize:]
	}
	return nil
}

type header struct {
	ChunkID   [4]byte // "RIFF"
	ChunkSize uint32  // 4 + (8 + format.ChunkSize) + (8 + data.ChunkSize)
	Format    [4]byte // "WAVE"
}

func (chunk *header) Read(data []byte) (rest []byte) {
	p := 0
	p += copy(chunk.ChunkID[:], data[p:])
	p += readU32LE(&chunk.ChunkSize, data[p:])
	p += copy(chunk.Format[:], data[p:])
	return data[p:]
}

type format struct {
	ChunkID       [4]byte // "fmt "
	ChunkSize     uint32  // 16 for PCM, size rest of header
	Encoding      encodings.Encoding
	NumChannels   uint16 // 1, 2, ...
	SampleRate    uint32 // 8000, 41000 ...
	ByteRate      uint32 // SampleRate * NumChannels * BitsPerSample / 8
	BlockAlign    uint16 // NumChannels * BitsPerSample / 8
	BitsPerSample uint16 // 8, 16
	// Extra parameters.
	ExSize          uint16
	ExBitsPerSample uint16
	ExChannelMask   uint32
	ExGUID          [16]byte
}

func (chunk *format) Read(data []byte) (rest []byte) {
	data = findSubChunk(data, [4]byte{'f', 'm', 't', ' '})
	p := 0

	p += copy(chunk.ChunkID[:], data[p:])
	p += readU32LE(&chunk.ChunkSize, data[p:])
	var af uint16 // AudioFormat
	p += readU16LE(&af, data[p:])
	chunk.Encoding = encodings.Encoding(af)
	p += readU16LE(&chunk.NumChannels, data[p:])
	p += readU32LE(&chunk.SampleRate, data[p:])
	p += readU32LE(&chunk.ByteRate, data[p:])
	p += readU16LE(&chunk.BlockAlign, data[p:])
	p += readU16LE(&chunk.BitsPerSample, data[p:])

	if chunk.Encoding != encodings.PCM {
		p += readU16LE(&chunk.ExSize, data[p:])
		p += readU16LE(&chunk.ExBitsPerSample, data[p:])
		p += readU32LE(&chunk.ExChannelMask, data[p:])
		p += copy(chunk.ExGUID[:], data[p:])
	}

	return data[8+int(chunk.ChunkSize):]
}

type data struct {
	ChunkID   [4]byte // "data"
	ChunkSize uint32  // NumSamples * NumChannels * BitsPerSample / 8. bytes in data
	Data      []byte
}

func (chunk *data) Read(data []byte) (rest []byte) {
	data = findSubChunk(data, [4]byte{'d', 'a', 't', 'a'})

	p := 0
	p += copy(chunk.ChunkID[:], data[p:])
	p += readU32LE(&chunk.ChunkSize, data[p:])

	k := p + int(chunk.ChunkSize)
	if k > len(data) {
		k = len(data)
	}
	chunk.Data = data[p:k]
	return data[k:]
}

func readU32LE(r *uint32, v []byte) int {
	*r = uint32(v[0])<<0 | uint32(v[1])<<8 | uint32(v[2])<<16 | uint32(v[3])<<24
	return 4
}

func readU16LE(r *uint16, v []byte) int {
	*r = uint16(v[0])<<0 | uint16(v[1])<<8
	return 2
}

func readU8LE(r *uint8, v []byte) int {
	*r = uint8(v[0])
	return 1
}
