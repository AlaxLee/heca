package heca

import (
	"time"
	"encoding/json"
	"strings"
	"net/http"
	"io/ioutil"
	log "github.com/cihub/seelog"
)

const (
	PUSH_INTERVAL = 1 * time.Millisecond   //单位毫秒
)

func SendToArgus(inChan chan interface{}) {

	pushLimit := Config().Push.Limit
	pushTimewait := Config().Push.TimeWait

	dataList := make([]interface{}, pushLimit)
	startTime := time.Now().Unix()
	i := 0

SEND:
	for {
		NowTime := time.Now().Unix()
		if i >= pushLimit ||  NowTime - startTime > pushTimewait {
			pushData( dataList[0:i] )
			dataList = make([]interface{}, pushLimit)
			startTime = NowTime
			i = 0
		}

		select {
		case data, ok := <-inChan:
			if ok {
				dataList[i] = data
				i++
			} else {
				if i > 0 {
					pushData( dataList[0:i] )
				}
				break SEND
			}
		default:
			time.Sleep(PUSH_INTERVAL)   //降低CPU使用率
		}
	}
}


func pushData(dataList []interface{}) {

	if len(dataList) == 0 {
		return
	}

	log.Debugf("now start push %d metrics", len(dataList))

	pushRetry := Config().Push.Retry

	for i := 0; i < pushRetry; i++ {
		err := pushOnceData(dataList)
		if err != nil {
			log.Warnf("try %i times failed: %s", i+1, err.Error())
			if i == pushRetry - 1 {
				log.Errorf("Final failed in %d times!!!", i+1)
			}
		} else {
			break
		}
	}
}

func pushOnceData(dataList []interface{}) error{
	APIUrl := Config().Push.APIUrl

	jsonBytes, err := json.Marshal(dataList)
	if err != nil {
		return err
	}

	resp, err := http.Post(APIUrl, "", strings.NewReader(string(jsonBytes)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	log.Debugf("push %d metrics to argus: %s" , len(dataList), string(body))

	return nil
}