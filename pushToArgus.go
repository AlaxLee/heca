package heca

import (
	"time"
	"encoding/json"
	"strings"
	"net/http"
	"io/ioutil"
	log "github.com/cihub/seelog"
)


const PUSH_LIMIT = 1000
const PUSH_TIMEWAIT = 10
const PUSH_RETRY = 3
const LOCAL_ARGUS_URL = "http://127.0.0.1:12050/v1/push"


func SendToArgus(inChan chan interface{}) {

	dataList := make([]interface{}, PUSH_LIMIT)
	startTime := time.Now().Unix()
	i := 0

	for {
		if i >= PUSH_LIMIT  ||  time.Now().Unix() - startTime > PUSH_TIMEWAIT {
			pushData( dataList[0:i] )
			dataList = make([]interface{}, PUSH_LIMIT)
			startTime = time.Now().Unix()
			i = 0
		}

		data, ok := <-inChan
		if ok {
			dataList[i] = data
			i++
		} else {
			if i > 0 {
				pushData( dataList[0:i] )
			}
			break
		}
	}
}


func pushData(dataList []interface{}) {
	for i := 0; i < PUSH_RETRY; i++ {
		err := pushOnceData(dataList)
		if err != nil {
			log.Warnf("try %i times failed: %s", i+1, err.Error())
			if i == PUSH_RETRY - 1 {
				log.Errorf("Final failed in %d times!!!", i+1)
			}
		} else {
			break
		}
	}
}

func pushOnceData(dataList []interface{}) error{
/*
	for _, d := range dataList {
		fmt.Print("<")
		lala, err := json.Marshal(d)
		if err != nil {
			log.Error(err)
		} else {
			fmt.Print(string(lala))
		}
		fmt.Print(">\n")
	}
	
	return nil
*/

	jsonBytes, err := json.Marshal(dataList)

	if err != nil {
		return err
	}

	resp, err := http.Post(LOCAL_ARGUS_URL, "", strings.NewReader(string(jsonBytes)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	log.Debug("push to argus: " + string(body))

	return nil
}


/*
#!-*- coding:utf8 -*-

import requests
import time
import json

ts = int(time.time())
payload = [
    {
        "endpoint": "test-endpoint",
        "metric": "test-metric",
        "timestamp": ts,
        "step": 60,
        "value": 1,
        "counterType": "GAUGE",
        "tags": "idc=lg,loc=beijing",
    },

    {
        "endpoint": "test-endpoint",
        "metric": "test-metric2",
        "timestamp": ts,
        "step": 60,
        "value": 2,
        "counterType": "GAUGE",
        "tags": "idc=lg,loc=beijing",
    },
]

r = requests.post("http://127.0.0.1:1988/v1/push", data=json.dumps(payload))

print r.text
*/
