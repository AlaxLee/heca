package heca

import (
	"time"
	"github.com/spf13/viper"
	log "github.com/cihub/seelog"
	"fmt"
	"errors"
)

const (
	JOB_STATUS_AVAILABLE  = "available"
	JOB_STATUS_RUNNING    = "running"
	JOB_STATUS_STOPPED    = "stopped"

	JOB_SIGNAL_CACHE_SIZE = 1
	JOB_SIGNAL_STOP       = "exit"

	JOB_EXCUTE_POOL_SIZE  = 3
)

type Job struct {
	id              string
	status          string             //available, running, stopped

	jobType         string             //这个值用来做executiveEntity的类型标识
	jobInterval     int                //执行周期，小于0或等于0 表示无周期

	starttime       time.Time          //任务的开始时间
	jobSignal       chan string        //任务的控制信息
	excutePool      chan int           //用于控制任务最多同时存在的个数

	result          chan<- interface{} //任务的结果反馈

	monConfigString string             //本job的配置，未曾解析过
	monConfig       *viper.Viper       //本job的配置，已经解析过了
	executiveEntity plugin             //具体执行体
}

func NewJob(id string, v *viper.Viper, resultQueue chan<- interface{}) (*Job, error) {

	jobType := v.GetString("jobType")
	if jobType == "" {
		return nil, errors.New("ERROR: " + id + "'s jobType is null")
	}

	jobInterval := v.GetInt("jobInterval")

	plugin, err := NewPlugin(jobType, v)
	if err != nil {
		return nil, err
	}

	configString, err := RearrangeJson(v)
	if err != nil {
		return nil, err
	}

	j := &Job{
		id:              id,
		status:          JOB_STATUS_AVAILABLE,
		jobType:         jobType,
		jobInterval:     jobInterval,
		starttime:       time.Now(),
		jobSignal:       make(chan string, JOB_SIGNAL_CACHE_SIZE),
		excutePool:      make(chan int, JOB_EXCUTE_POOL_SIZE),
		result:          resultQueue,
		monConfigString: configString,
		monConfig:       v,
		executiveEntity: plugin,
	}

	return j, err
}


//stop 并不会立即结束j，它会等待 time.Sleep(jobInterval)
func (j *Job)stop(){
	j.jobSignal <- JOB_SIGNAL_STOP
}


func (j *Job)start() {
	defer func() {
		j.status = JOB_STATUS_STOPPED
	}()
	j.status = JOB_STATUS_RUNNING

	Loop:
	for {
		select {
		case m := <- j.jobSignal:
			switch m {
			case JOB_SIGNAL_STOP:
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
		"config" : j.monConfig.AllSettings(),
	}
	return stat
}

