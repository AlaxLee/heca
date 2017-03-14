package heca

import (
	"time"
	"github.com/spf13/viper"
	log "github.com/cihub/seelog"
	"bytes"
	"errors"
	"sync"
	"fmt"
)


type Controller struct {
	Total                   uint                   //当前注册的 controller 总数量
	Seq                     uint                   //本 controller 的编号

	resultQueue             chan interface{}       //用于存放执行结果的队列

	jobsMonConfigMutex      *sync.RWMutex
	jobsMonConfig           map[string]interface{} //这里 map 的 key 是 Job.id，用于快速比对新老job是否一致
	jobsMonConfigChangeTime time.Time

	jobsMutex               *sync.RWMutex
	jobs                    map[string]*Job        //这里 map 的 key 是 Job.id
}

func NewController() *Controller {

	ct := &Controller{

		resultQueue:              make(chan interface{}, Config().Controller.ResultQueueLength),

		jobsMonConfigMutex:       &sync.RWMutex{},
		jobsMonConfig:            make(map[string]interface{}),
		jobsMonConfigChangeTime:  time.Now(),

		jobsMutex:                &sync.RWMutex{},
		jobs:                     make(map[string]*Job),
	}

	return ct
}


func (ct *Controller) Start() {
	//startWatcher  启动一个监听，用于监控配置的变化，并根据监控的变化来起停job
	ct.startWatcher()

	//把执行结果处理掉，一般是发送到其他地方，或者持久化
	ct.dealResult()

	//启API监听
	ct.startAPI()

	//保持主进程活着
	select {}
}



func (ct *Controller) startWatcher() {
	watcherInterval := time.Duration(Config().Controller.WatcherInterval) * time.Second
	watcherEnabled  := Config().Controller.WatcherEnabled
	go func() {
		ct.reloadAllJobs()
		for {
			time.Sleep( watcherInterval )
			if watcherEnabled {
				ct.reloadAllJobs()
			}
		}
	}()
}

func (ct *Controller) dealResult() {
	go SendToArgus(ct.resultQueue)
}

func (ct *Controller) startAPI() {
	NewApiServer(ct).start()
}




func (ct *Controller) reloadAllJobs() {
	total, seq, jobsMonConfig := getJobMonConfig()

	//对比当前的 ct.jobsConfig 和 新得到的 jogConfigs，然后新增的Add，有变化的Update，减少的Del，未变化的检查running状态
	addList := make(map[string]string)
	updateList := make(map[string]string)
	delList := make(map[string]string)
	checkList := make(map[string]string)

	//根据 add、update、del、check 列表分别增、改、删、检查job
	defer ct.jobsMutex.Unlock()
	ct.jobsMutex.Lock()


	for oldId := range ct.jobs {
		delList[oldId] = ""
	}

	for newId, newConfigString := range jobsMonConfig {
		job, ok := ct.jobs[newId]
		if ok {
			if newConfigString != job.monConfigString {
				updateList[newId] = newConfigString
			} else {
				checkList[newId] = ""
			}
			delete(delList, newId)
		} else {
			addList[newId] = newConfigString
			fmt.Println(newId)
		}
	}



	for id, configString := range addList {
		fmt.Println("Add: " + id)
		j, err := ct.createJobObj(id, configString)
		if err != nil {
			log.Errorf("ERROR: create %s's job failed\n", id, err.Error())
		} else {
			ct.jobs[id] = j
			go j.start()
		}
	}

	for id, configString := range updateList {
		fmt.Println("Update: " + id)
		j, err := ct.createJobObj(id, configString)
		if err != nil {
			log.Errorf("ERROR: create %s's job failed\n", id, err.Error())
		} else {
			ct.jobs[id].stop()
			ct.jobs[id] = j
			go j.start()
		}
	}

	for id := range delList {
		fmt.Println("Del: " + id)
		ct.jobs[id].stop()
		delete(ct.jobs, id)
	}

	for id := range checkList {
		fmt.Println("Check: " + id)
		j := ct.jobs[id]
		if j.jobInterval > 0  &&  j.status == "stopped" {
			go j.start()
		}
	}

	defer ct.jobsMonConfigMutex.Unlock()
	ct.jobsMonConfigMutex.Lock()

	ct.Total = total
	ct.Seq = seq

	ct.jobsMonConfig = make(map[string]interface{})
	for id, j := range ct.jobs {
		ct.jobsMonConfig[id] = j.monConfig.AllSettings()
	}
	ct.jobsMonConfigChangeTime = time.Now()
}





func (ct *Controller) createJobObj(id string, configString string) (*Job, error) {

	v := viper.New()

	v.SetConfigType("json")

	err := v.ReadConfig(bytes.NewBuffer( []byte(configString) ))
	if err != nil {
		return nil, errors.New("ERROR: parse " + id + "'s configString failed: " + err.Error())
	}

	j, err := NewJob(id, v, ct.resultQueue)
	if err != nil {
		return nil, errors.New("ERROR: create job [" + id + "]'s object failed: " + err.Error())
	}

	return j, nil
}



func (ct *Controller) GetInstanceTotal() uint {
	return ct.Total
}

func (ct *Controller) GetInstanceSeq() uint {
	return ct.Seq
}

func (ct *Controller) GetConfig() map[string]interface{} {

	defer ct.jobsMonConfigMutex.RUnlock()
	ct.jobsMonConfigMutex.RLock()
	return ct.jobsMonConfig
}

func (ct *Controller) AddJob(id string, originConfigString string)  (map[string]interface{}, error) {

	configString, err := RearrangeJson(originConfigString)
	if err != nil {
		log.Errorf("ERROR: rearrange %s's config failed: %s\n", id, err.Error())
		return nil, errors.New(fmt.Sprintf("ERROR: rearrange %s's config failed: %s\n", id, err.Error()))
	}

	j, err := ct.createJobObj(id, configString)
	if err != nil {
		log.Errorf("ERROR: create %s's job failed: %s\n", id, err.Error())
		return nil, errors.New(fmt.Sprintf("ERROR: create %s's job failed: %s\n", id, err.Error()))
	}

	defer ct.jobsMutex.Unlock()
	ct.jobsMutex.Lock()

	job, ok := ct.jobs[id]

	if ok {
		if job.monConfigString != configString {
			log.Errorf("job[" + id + "] has exists，but is not equal: \n" + job.monConfigString + "\n" + configString + "\n" )
			return nil, errors.New("job[" + id + "] has exists，but config is not equal")
		} else {
			return job.getCurrentStat(), nil
		}
	} else {
		ct.jobs[id] = j
		go j.start()

		defer ct.jobsMonConfigMutex.Unlock()
		ct.jobsMonConfigMutex.Lock()

		ct.jobsMonConfig[id] = j.monConfig.AllSettings()
		ct.jobsMonConfigChangeTime = time.Now()

		return j.getCurrentStat(), nil

	}
}

func (ct *Controller) DelJob(id string) (map[string]interface{}, error) {
	defer ct.jobsMonConfigMutex.Unlock()
	ct.jobsMonConfigMutex.Lock()
	defer ct.jobsMutex.Unlock()
	ct.jobsMutex.Lock()

	job, ok := ct.jobs[id]

	var result map[string]interface{}

	if ok {
		job.stop()
		result = job.getCurrentStat()
		delete(ct.jobsMonConfig, id)
		delete(ct.jobs, id)

	} else {
		result = make(map[string]interface{})
	}

	return result, nil
}

func (ct *Controller) UpdateJob(id string, originConfigString string) (map[string]interface{}, error) {

	configString, err := RearrangeJson(originConfigString)
	if err != nil {
		log.Errorf("ERROR: rearrange %s's config failed: %s\n", id, err.Error())
		return nil, errors.New(fmt.Sprintf("ERROR: rearrange %s's config failed: %s\n", id, err.Error()))
	}


	j, err := ct.createJobObj(id, configString)
	if err != nil {
		log.Errorf("ERROR: create %s's job failed: %s\n", id, err.Error())
		return nil, errors.New(fmt.Sprintf("ERROR: create %s's job failed: %s\n", id, err.Error()))
	}

	defer ct.jobsMutex.Unlock()
	ct.jobsMutex.Lock()

	job, ok := ct.jobs[id]

	if !ok {
		log.Errorf("job[" + id + "] has not exists")
		return nil, errors.New("job[" + id + "] has not exists")
	}



	if job.monConfigString != configString {
		job.stop()
		ct.jobs[id] = j
		go j.start()

		defer ct.jobsMonConfigMutex.Unlock()
		ct.jobsMonConfigMutex.Lock()

		ct.jobsMonConfig[id] = j.monConfig.AllSettings()
		ct.jobsMonConfigChangeTime = time.Now()

		return j.getCurrentStat(), nil

	} else {
		return job.getCurrentStat(), nil
	}

}

func (ct *Controller) GetJob(id string) (result map[string]interface{}) {

	result = make(map[string]interface{})
	if j, ok := ct.jobs[id]; ok {
		result[id] = j.getCurrentStat()
	}
	return result
}

func (ct *Controller) GetAllJob() (result map[string]interface{}) {

	result = make(map[string]interface{})
	for id, j := range ct.jobs {
		result[id] = j.getCurrentStat()
	}
	return result
}


func (ct *Controller) ReloadAllJobs() map[string]interface{} {
	ct.reloadAllJobs()
	return ct.GetAllJob()
}


