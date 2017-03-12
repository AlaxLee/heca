package heca

import (
	"path/filepath"
	"os"
)






type PushConfig struct {
}


type GlobalConfig struct {
}



var (
	Home         string
	globalConfig *GlobalConfig
)

func GetGlobalConfig() *GlobalConfig {

	return globalConfig
}


func init() {
	home, err := filepath.Abs(filepath.Dir(os.Args[0]) + "/../")

	if err != nil {
		panic(err)
	}
	Home = home
}
