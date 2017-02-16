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
	configs map[string]interface{}   //这里 map 的 key 是 Job.id
	changeTime time.Time
}

type Job struct {
	id              string
	status          string       //avaliable, running, stopped

	jobType         string       //这个值用来做JobDo的类型名字
	jobInterval     int          //执行周期，小于0或等于0 表示无周期

	starttime       time.Time
	mesg            chan string

	configString    string       //本job的配置，未曾解析过
	config          *viper.Viper //本job的配置，已经解析过了
	executiveEntity plugin       //具体执行体
}


func NewController() *Controller {

	ct := &Controller{
		config: Config{
			mux: &sync.RWMutex{},
			configs: make(map[string]interface{}),
			changeTime: time.Now(),
		},
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
	go func() {
		fmt.Println(<- ct.resultQueue)
	}()
}

func (ct *Controller) startAPI() {
	NewApiServer(ct).start()
}




func (ct *Controller) reloadAllJobs() {
	total, seq, configs := getConfig()

	//对比当前的 ct.configs 和 新得到的 configs，然后新增的Add，有变化的Update，减少的Del，未变化的检查running状态
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

	for newId, newConfigString := range configs {
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
		j, err := createJobObj(id, configString)
		if err != nil {
			log.Errorf("ERROR: create %s's job failed\n", id, err.Error())
		} else {
			ct.jobs[id] = j
			go startJob(j, ct.resultQueue)
		}
	}

	for id, configString := range updateList {
		fmt.Println("Update: " + id)
		j, err := createJobObj(id, configString)
		if err != nil {
			log.Errorf("ERROR: create %s's job failed\n", id, err.Error())
		} else {
			stopJob( ct.jobs[id] )
			ct.jobs[id] = j
			go startJob(j, ct.resultQueue)
		}
	}

	for id := range delList {
		fmt.Println("Del: " + id)
		stopJob( ct.jobs[id] )
		delete(ct.jobs, id)
	}

	for id := range checkList {
		fmt.Println("Check: " + id)
		j := ct.jobs[id]
		if j.jobInterval > 0  &&  j.status == "stopped" {
			go startJob(j, ct.resultQueue)
		}
	}

	defer ct.config.mux.Unlock()
	ct.config.mux.Lock()

	ct.config.total = total
	ct.config.seq = seq

	ct.config.configs = make(map[string]interface{})
	for id, j := range ct.jobs {
		ct.config.configs[id] = j.config.AllSettings()
		//err := json.Unmarshal([]bytes(j.configString), ct.config.configs[id])
		//if err != nil {
		//	log.Errorf("ERROR: unmarshal %s's job configString failed\n", j.id, err.Error())
		//	continue
		//}
	}
	ct.config.changeTime = time.Now()
}





func createJobObj(id string, configString string) (*Job, error) {

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
		configString: configString,
		config: v,
		executiveEntity: plugin,
	}

	return j, nil
}



//stopJob 并不会立即结束j，它会等待 time.Sleep(jobInterval)：
func stopJob(j *Job) {
	j.mesg <- "exit"
}

func startJob(j *Job, resultQueue chan interface{}) {
	defer func() {
		j.status = "stopped"
	}()
	j.status = "running"

	excutePool := make(chan int, 3)
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
		go func(job *Job, exeP chan int) {
			var isRunning bool
			defer func(ep chan int) {
				if isRunning {
					select {
					case <-ep:
					default:
						log.Error("ERROR: this should not happened")
					}
				}
			}(exeP)

			select {
			case exeP <- 1:
				isRunning = true
				//job.executiveEntity.Do()
				resultQueue <- job.executiveEntity.Do()
			default:
				isRunning = false
			}
		}(j, excutePool)

		if j.jobInterval > 0 {
			time.Sleep( time.Duration(j.jobInterval)*time.Second )
		} else {
			break
		}
	}
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
	return ct.config.configs
}

func (ct *Controller) AddJob(id string, configString string)  (interface{}, error) {

	defer ct.jobsMutex.Unlock()
	ct.jobsMutex.Lock()

	job, ok := ct.jobs[id]

	if ok {
		//这里的判断其实是过于严厉了
		//例如：两个json string，只是顺序不一样，这样配置实际上是一样的，但这里会被认为是不一样
		if job.configString == configString {
			return job.config.AllSettings(), nil
		} else {
			log.Errorf("job[" + id + "] has exists，but config is not equal：\n" + "old:\n" + job.configString + "\nnew:\n" + configString)
			return nil, errors.New("job[" + id + "] has exists，but config is not equal：\n" + "old:\n" + job.configString + "\nnew:\n" + configString)
		}
	} else {
		j, err := createJobObj(id, configString)
		if err != nil {
			log.Errorf("ERROR: create %s's job failed: %s\n", id, err.Error())
			return nil, errors.New(fmt.Sprintf("ERROR: create %s's job failed: %s\n", id, err.Error()))
		} else {
			ct.jobs[id] = j
			go startJob(j, ct.resultQueue)
		}

		defer ct.config.mux.Unlock()
		ct.config.mux.Lock()

		ct.config.configs[id] = j.config.AllSettings()
		ct.config.changeTime = time.Now()

		return ct.config.configs[id], nil
	}
}

func (ct *Controller) DelJob(id string) (interface{}, error) {
	defer ct.config.mux.Unlock()
	ct.config.mux.Lock()
	defer ct.jobsMutex.Unlock()
	ct.jobsMutex.Lock()

	job, ok := ct.jobs[id]

	var result interface{}

	if ok {
		stopJob(job)
		result = job.config.AllSettings()
		delete(ct.config.configs, id)
		delete(ct.jobs, id)

	}

	return result, nil
}

func (ct *Controller) UpdateJob(id string, configString string) (interface{}, error) {

	defer ct.jobsMutex.Unlock()
	ct.jobsMutex.Lock()

	job, ok := ct.jobs[id]

	if ok {
		if job.configString == configString {
			return job.config.AllSettings(), nil
		} else {
			j, err := createJobObj(id, configString)
			if err != nil {
				log.Errorf("ERROR: create %s's job failed: %s\n", id, err.Error())
				return nil, errors.New(fmt.Sprintf("ERROR: create %s's job failed: %s\n", id, err.Error()))
			} else {
				stopJob( ct.jobs[id] )
				ct.jobs[id] = j
				go startJob(j, ct.resultQueue)
			}

			defer ct.config.mux.Unlock()
			ct.config.mux.Lock()

			ct.config.configs[id] = j.config.AllSettings()
			ct.config.changeTime = time.Now()

			return ct.config.configs[id], nil

		}
	} else {
		log.Errorf("job[" + id + "] has not exists")
		return nil, errors.New("job[" + id + "] has not exists")
	}
}

func (ct *Controller) GetJob(id string) interface{} {
	defer ct.config.mux.RUnlock()
	ct.config.mux.RLock()
	return ct.config.configs[id]
}

func (ct *Controller) GetAllJob() map[string]interface{} {
	defer ct.config.mux.RUnlock()
	ct.config.mux.RLock()
	return ct.config.configs
}


func (ct *Controller) ReloadAllJobs() map[string]interface{} {
	ct.reloadAllJobs()
	return ct.config.configs
}


