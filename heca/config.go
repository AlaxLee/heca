package heca

import (
	"io/ioutil"
	"log"
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
		log.Fatal(err)
	}

	configs = make(map[string]string)
	
	for _, file := range files {
		//fmt.Println(file.Name())
		//fmt.Printf("%T\n", file)

		filePath := dir + "/" + file.Name()

		result, err := ioutil.ReadFile(filePath)

		if err != nil {
			panic(err)
		}

		//fmt.Println(string(result))

		configs[file.Name()] = string(result)
	}

	return
}