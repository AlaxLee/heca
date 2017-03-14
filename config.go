package heca

import (
	log "github.com/cihub/seelog"
	"fmt"
	"encoding/json"
	"errors"
	"regexp"
	"os"
	"bufio"
	"github.com/spf13/viper"
	"io/ioutil"
	"sync"
	"path/filepath"
)



type ControllerConfig struct {
	ResultQueueLength  int   `json:"resultQueueLength"`     //结果存放队列
	WatcherEnabled     bool  `json:"watcherEnabled"`        //是否使自动监听生效
	WatcherInterval    int   `json:"watcherInterval"`       //自动监听的检查的时间间隔
								//Total int  `json:"total"`
								//Seq   int  `json:"seq"`
}

type PushConfig struct {
	Limit     int    `json:"limit"`     //推送一次的最大数量
	TimeWait  int64  `json:"timewait"`  //推送一次的时间间隔
	Retry     int    `json:"retry"`     //推送的重试次数
	APIUrl    string `json:"apiUrl"`    //推送的接口
}


type GlobalConfig struct {
	Push       *PushConfig        `json:"push"`
	Controller *ControllerConfig  `json:"controller"`
}


var (
	Home              string
	globalConfig      *GlobalConfig
	globalConfigLock  = new(sync.RWMutex)
)

func Config() *GlobalConfig {

	defer globalConfigLock.RUnlock()
	globalConfigLock.RLock()

	return globalConfig
}







func init() {
	home, err := filepath.Abs(filepath.Dir(os.Args[0]) + "/../")

	if err != nil {
		panic(err)
	}
	Home = home
}

func InitConfig()  error{
	return ParseConfig(Home + "/conf/heca.json")
}

func ParseConfig(configFilePath string) error {
	configContent, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return err
	}

	var c GlobalConfig
	err = json.Unmarshal(configContent, &c)
	if err != nil {
		return err
	}

	globalConfigLock.Lock()
	defer globalConfigLock.Unlock()

	globalConfig = &c

	log.Info("read config file: ", configFilePath, " successfully")
	return nil
}




func getJobConfig() (total uint, seq uint, configs map[string]string){

	total = 2
	seq = 1

	//configs = map[string]string {
	//	"baidu" : `
	//	{
	//		"jobType": "ping",
	//		"jobInterval" : 15,
	//		"address": "www.baidu.com",
	//		"timeout": 3,
	//		"retry": 3
	//	}`,
	//	"qidian" : `
	//	{
	//		"jobType": "ping",
	//		"jobInterval" : 15,
	//		"address": "www.qidian.com",
	//		"timeout": 3,
	//		"retry": 3
	//	}`,
	//}

	configs = make(map[string]string)


	configFilePath := Home + "/conf/hostlist"

	blankRegx, err := regexp.Compile("\\s+")
	if err != nil {
		log.Error(err)
		return
	}


	f, err := os.Open(configFilePath)
	if err != nil {
		log.Error(err)
		return
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		info := blankRegx.Split(s.Text(), 2)
		if len(info) != 2 {
			log.Info(info)
			continue
		}
		jobId := info[0]
		address := info[1]

		jobInfo := map[string]interface{} {
			"endpoint": jobId,
			"jobType": "ping",
			"jobInterval" : 60,
			"address": address,
			"timeout": 3,
			"retry": 3,
		}

		configContent, err := RearrangeJson(jobInfo)
		if err != nil {
			log.Error(err)
			continue
		}

		configs[jobId] = configContent

	}
	return
}


func RearrangeJson(i interface{})(rearrangedJsonString string, err error) {

	var rearrangedJsonBytes []byte

	switch inputObject := i.(type) {
	case *viper.Viper:
		rearrangedJsonBytes, err = json.Marshal(inputObject.AllSettings())
	case map[string]interface{}:
		rearrangedJsonBytes, err = json.Marshal(inputObject)
	case string:
		tmpMap := make(map[string]interface{})
		err = json.Unmarshal([]byte(inputObject), &tmpMap)
		if err == nil {
			rearrangedJsonBytes, err = json.Marshal(tmpMap)
		}
	case []byte:
		tmpMap := make(map[string]interface{})
		err = json.Unmarshal(inputObject, &tmpMap)
		if err == nil {
			rearrangedJsonBytes, err = json.Marshal(tmpMap)
		}
	default:
		err = errors.New(fmt.Sprintf("type %T is not supported", inputObject))
	}

	if rearrangedJsonBytes != nil {
		rearrangedJsonString = string(rearrangedJsonBytes)
	}

	return
}
