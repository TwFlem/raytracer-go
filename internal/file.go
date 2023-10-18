package internal

import (
	"image"
	"image/jpeg"
	"os"
)

func Overwrite(fname string) (*os.File, error) {
	if _, err := os.Stat(fname); err == nil {
		if err = os.Remove(fname); err != nil {
			return nil, err
		}
	}
	f, err := os.Create(fname)
	return f, err

}

func LoadJPEG(fname string) (image.Image, error) {
	f, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return jpeg.Decode(f)
}
