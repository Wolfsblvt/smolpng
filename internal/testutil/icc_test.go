package testutil

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"io"
	"strings"
	"testing"

	"github.com/Wolfsblvt/smolpng/internal/jpegimage"
	"github.com/Wolfsblvt/smolpng/internal/pngimage"
)

func profileTag(t *testing.T, profile []byte, name string) []byte {
	t.Helper()
	count := binary.BigEndian.Uint32(profile[128:])
	for i := uint32(0); i < count; i++ {
		entry := profile[132+i*12:]
		if string(entry[:4]) == name {
			offset, size := binary.BigEndian.Uint32(entry[4:]), binary.BigEndian.Uint32(entry[8:])
			return profile[offset : offset+size]
		}
	}
	t.Fatalf("missing profile tag %s", name)
	return nil
}

func TestICCControlHasACompleteRGBProfileAndDifferentInterpretation(t *testing.T) {
	standard, swapped := ICCProfile(false), ICCProfile(true)
	for _, profile := range [][]byte{standard, swapped} {
		if int(binary.BigEndian.Uint32(profile)) != len(profile) || string(profile[36:40]) != "acsp" || string(profile[16:20]) != "RGB " {
			t.Fatal("fixture is not a complete RGB ICC profile")
		}
		if len(profileTag(t, profile, "rXYZ")) != 20 || len(profileTag(t, profile, "bXYZ")) != 20 {
			t.Fatal("control needs RGB matrix colorants")
		}
	}
	if bytes.Equal(standard, swapped) || !bytes.Equal(profileTag(t, standard, "rXYZ"), profileTag(t, swapped, "bXYZ")) || !bytes.Equal(profileTag(t, standard, "bXYZ"), profileTag(t, swapped, "rXYZ")) {
		t.Fatal("swapped control lost its different color interpretation")
	}
}

func TestICCCorpusCarriesCompleteProfilesThroughSafeMetadata(t *testing.T) {
	for name, data := range ICCCorpus() {
		t.Run(name, func(t *testing.T) {
			profile := ICCProfile(strings.Contains(name, "swapped"))
			if strings.HasSuffix(name, ".jpg") {
				before, err := jpegimage.Decode(data, 1_000_000)
				if err != nil {
					t.Fatal(err)
				}
				want := jpegimage.Marker(0xe2, append([]byte("ICC_PROFILE\x00\x01\x01"), profile...))
				if !bytes.Equal(before.ICC(), want) {
					t.Fatal("JPEG ICC sequence is incomplete")
				}
				safe, err := jpegimage.Filter(before, "safe")
				if err != nil {
					t.Fatal(err)
				}
				after, err := jpegimage.Decode(safe, 1_000_000)
				if err != nil || !bytes.Equal(after.ICC(), want) || !pngimage.PixelEqual(before.Image, after.Image) {
					t.Fatal("safe metadata lost the JPEG raster/profile", err)
				}
				return
			}
			chunks, err := pngimage.Chunks(data)
			if err != nil {
				t.Fatal(err)
			}
			var carrier []byte
			for _, chunk := range chunks {
				if chunk.Type != "iCCP" {
					continue
				}
				carrier = chunk.Data
				compressed := bytes.SplitN(carrier, []byte{0}, 2)
				if len(compressed) != 2 || len(compressed[1]) < 2 || compressed[1][0] != 0 {
					t.Fatal("invalid PNG ICC carrier")
				}
				reader, err := zlib.NewReader(bytes.NewReader(compressed[1][1:]))
				if err != nil {
					t.Fatal(err)
				}
				decoded, err := io.ReadAll(reader)
				_ = reader.Close()
				if err != nil || !bytes.Equal(decoded, profile) {
					t.Fatal("PNG ICC payload is incomplete", err)
				}
			}
			if carrier == nil {
				t.Fatal("PNG is missing the real profile")
			}
			safe, err := pngimage.Filter(data, "safe")
			if err != nil || !bytes.Contains(safe, carrier) {
				t.Fatal("safe metadata lost the PNG profile", err)
			}
		})
	}
}
