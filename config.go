package heca

import (
	log "github.com/cihub/seelog"
	"fmt"
	"encoding/json"
	"errors"
	"regexp"
	"os"
	"bufio"
)



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
