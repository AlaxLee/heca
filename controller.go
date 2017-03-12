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
	config Config   //只起记录作用
	enableWatcher bool //是否使自动监听生效
	resultQueue chan interface{} //用于存放执行结果的队列


	jobsMutex *sync.RWMutex
	jobs map[string]*Job    //这里 map 的 key 是 Job.id
}

type Config struct {
	total uint	//当前注册的 controller 总数量
	seq uint	//本 controller 的编号

	mux *sync.RWMutex
	jobConfigs map[string]interface{}   //这里 map 的 key 是 Job.id，用于快速比对新老job是否一致
	changeTime time.Time
}

type Job struct {
	id              string
	status          string       //avaliable, running, stopped

	jobType         string       //这个值用来做executiveEntity的类型标识
	jobInterval     int          //执行周期，小于0或等于0 表示无周期

	starttime       time.Time    //任务的开始时间
	mesg            chan string  //任务的控制信息
	excutePool	chan int     //用于控制任务最多同时存在的个数

	result		chan<- interface{}  //任务的结果反馈

	configString    string       //本job的配置，未曾解析过
	config          *viper.Viper //本job的配置，已经解析过了
	executiveEntity plugin       //具体执行体
}


func NewController() *Controller {

	ct := &Controller{
		config: Config{
			mux: &sync.RWMutex{},
			jobConfigs: make(map[string]interface{}),
			changeTime: time.Now(),
		},
		resultQueue: make(chan interface{}, 5 * 1000 * 1000),
		jobsMutex: &sync.RWMutex{},
		jobs: make(map[string]*Job),
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
	for {
		time.Sleep( 5 * time.Second )
	}
}



func (ct *Controller) startWatcher() {
	go func() {
		ct.reloadAllJobs()
		for {
			time.Sleep( 60 * time.Second )
			if ct.enableWatcher {
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
	total, seq, jobConfigs:= getJobConfig()

	//对比当前的 ct.jobConfigs 和 新得到的 jogConfigs，然后新增的Add，有变化的Update，减少的Del，未变化的检查running状态
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

	for newId, newConfigString := range jobConfigs {
		job, ok := ct.jobs[newId]
		if ok {
			if newConfigString != job.configString {
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

	defer ct.config.mux.Unlock()
	ct.config.mux.Lock()

	ct.config.total = total
	ct.config.seq = seq

	ct.config.jobConfigs = make(map[string]interface{})
	for id, j := range ct.jobs {
		ct.config.jobConfigs[id] = j.config.AllSettings()
		//err := json.Unmarshal([]bytes(j.configString), ct.config.jobConfigs[id])
		//if err != nil {
		//	log.Errorf("ERROR: unmarshal %s's job configString failed\n", j.id, err.Error())
		//	continue
		//}
	}
	ct.config.changeTime = time.Now()
}





func (ct *Controller) createJobObj(id string, configString string) (*Job, error) {

	v := viper.New()

	v.SetConfigType("json")

	err := v.ReadConfig(bytes.NewBuffer( []byte(configString) ))

	if err != nil {
		return nil, errors.New("ERROR: parse " + id + "'s configString failed: " + err.Error())
	}

	jobType := v.GetString("jobType")
	if jobType == "" {
		return nil, errors.New("ERROR: " + id + "'s jobType is null")
	}

	jobInterval := v.GetInt("jobInterval")

	plugin, err := NewPlugin(jobType, v)
	if err != nil {
		return nil, err
	}

	j := &Job{
		id: id,
		status: "avaliable",
		jobType: jobType,
		jobInterval: jobInterval,
		starttime: time.Now(),
		mesg: make(chan string, 1),
		excutePool: make(chan int, 3),
		result: ct.resultQueue,
		configString: configString,
		config: v,
		executiveEntity: plugin,
	}

	return j, nil
}



//stop 并不会立即结束j，它会等待 time.Sleep(jobInterval)
func (j *Job)stop(){
	j.mesg <- "exit"
}


func (j *Job)start() {
	defer func() {
		j.status = "stopped"
	}()
	j.status = "running"

	Loop:
	for {
		select {
		case m := <- j.mesg:
			switch m {
			case "exit":
				fmt.Println("job " + j.id + " get stop signal")
				break Loop
			}
		default:
		}

		//之所以这样单独起一个 goroutine ，是为了保证时间间隔的准确性
		//但是，如果这个 goroutine 不能迅速结束，那么 goroutine 会越来越多……
		//所以专门用了一个 chain 做限制
		go func(job *Job) {
			var isRunning bool
			defer func(ep chan int) {
				if isRunning {
					select {
					case <-ep:
					default:
						log.Error("ERROR: this should not happened")
					}
				}
			}(job.excutePool)

			select {
			case job.excutePool <- 1:
				isRunning = true
				job.executiveEntity.Do(j.result)
/*
				if job.result == nil {
					job.executiveEntity.Do()
				} else {
					job.result <- job.executiveEntity.Do()
				}
*/
			default:
				job.status = "blocking"
				isRunning = false
			}
		}(j)

		if j.jobInterval > 0 {
			time.Sleep( time.Duration(j.jobInterval)*time.Second )
		} else {
			break
		}
	}
}

func (j *Job) getCurrentStat() (stat map[string]interface{}){

	stat = map[string]interface{} {
		"status" : j.status,
		"config" : j.config.AllSettings(),
	}
	return stat
}


func (ct *Controller) GetInstanceTotal() uint {
	return ct.config.total
}

func (ct *Controller) GetInstanceSeq() uint {
	return ct.config.seq
}

func (ct *Controller) GetConfig() map[string]interface{} {

	defer ct.config.mux.RUnlock()
	ct.config.mux.RLock()
	return ct.config.jobConfigs
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
		if job.configString != configString {
			log.Errorf("job[" + id + "] has exists，but is not equal: \n" + job.configString + "\n" + configString + "\n" )
			return nil, errors.New("job[" + id + "] has exists，but config is not equal")
		} else {
			return job.getCurrentStat(), nil
		}
	} else {
		ct.jobs[id] = j
		go j.start()

		defer ct.config.mux.Unlock()
		ct.config.mux.Lock()

		ct.config.jobConfigs[id] = j.config.AllSettings()
		ct.config.changeTime = time.Now()

		return j.getCurrentStat(), nil

	}
}

func (ct *Controller) DelJob(id string) (map[string]interface{}, error) {
	defer ct.config.mux.Unlock()
	ct.config.mux.Lock()
	defer ct.jobsMutex.Unlock()
	ct.jobsMutex.Lock()

	job, ok := ct.jobs[id]

	var result map[string]interface{}

	if ok {
		job.stop()
		result = job.getCurrentStat()
		delete(ct.config.jobConfigs, id)
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



	if job.configString != configString {
		job.stop()
		ct.jobs[id] = j
		go j.start()

		defer ct.config.mux.Unlock()
		ct.config.mux.Lock()

		ct.config.jobConfigs[id] = j.config.AllSettings()
		ct.config.changeTime = time.Now()

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


