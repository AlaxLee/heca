package heca

import (
	"io/ioutil"
	log "github.com/cihub/seelog"
	"fmt"
	"encoding/json"
	"errors"
)

func getConfig() (total uint, seq uint, configs map[string]string){

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

	dir := "/Users/user/IdeaProjects/Hecatoncheires/conf"

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Error(err)
		return
	}

	configs = make(map[string]string)
	
	for _, file := range files {
		//fmt.Println(file.Name())
		//fmt.Printf("%T\n", file)

		filePath := dir + "/" + file.Name()

		result, err := ioutil.ReadFile(filePath)

		if err != nil {
			log.Error(err)
			continue
		}

		//fmt.Println(string(result))
		configContent, err := RearrangeJson(result)
		if err != nil {
			log.Error(err)
			continue
		}
		configs[file.Name()] = configContent
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