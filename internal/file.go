package internal

import (
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
