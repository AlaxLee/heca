package heca

import (
	"path/filepath"
	"os"
)

var Home string

func init() {
	home, err := filepath.Abs(filepath.Dir(os.Args[0]) + "/../")

	if err != nil {
		panic(err)
	}
	Home = home
}