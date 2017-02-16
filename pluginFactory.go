package heca

import (
	"errors"
	"github.com/spf13/viper"
	myping "github.com/AlaxLee/heca/plugin/ping"
)
//func New(jobType string) (obj interface{}, err error) {
//	switch jobType {
//	case "ping":
//		obj := NewMyPing()
//	default:
//
//	}
//}

type plugin interface {
	Do() interface{}
}

func NewPlugin(pluginType string, config *viper.Viper) (plugin plugin, err error) {
	switch pluginType {
	case "ping":
		plugin, err = myping.NewMyPing(config)
	default:
		err = errors.New("ERROR: " + pluginType + "is not supported")
	}
	return plugin, err
}
