package heca

import (
	"fmt"
	"time"

	"net/http"
	_ "net/http/pprof"
	"encoding/json"
	"errors"
)

type apiServer struct {
	ct *Controller
}

func NewApiServer(controller *Controller) *apiServer {
	return &apiServer{ct: controller}
}

func (a *apiServer) start() {

	http.HandleFunc("/api/job/search", func(w http.ResponseWriter, r *http.Request) {
		jobid := r.FormValue("jobid")

		result, err := a.SearchJob(jobid)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		RenderJson(w, result)
	})


	http.HandleFunc("/api/job/searchall", func(w http.ResponseWriter, r *http.Request) {
		result, err := a.SearchAllJob()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		RenderJson(w, result)
	})


	http.HandleFunc("/api/job/delete", func(w http.ResponseWriter, r *http.Request) {
		jobid := r.FormValue("jobid")

		result, err := a.DeleteJob(jobid)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		RenderJson(w, result)
	})

	http.HandleFunc("/api/job/add", func(w http.ResponseWriter, r *http.Request) {
		jobid := r.FormValue("jobid")
		jobConfigString := r.FormValue("config")

		result, err := a.AddJob(jobid, jobConfigString)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		RenderJson(w, result)
	})

	http.HandleFunc("/api/job/update", func(w http.ResponseWriter, r *http.Request) {
		jobid := r.FormValue("jobid")
		jobConfigString := r.FormValue("config")

		result, err := a.UpdateJob(jobid, jobConfigString)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		RenderJson(w, result)
	})


	http.HandleFunc("/api/job/reloadall", func(w http.ResponseWriter, r *http.Request) {
		result, err := a.ReloadAllJob()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		RenderJson(w, result)
	})


	go func() {
		fmt.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	for {
		time.Sleep( 5 * time.Second )
	}
}

func RenderJson(w http.ResponseWriter, v interface{}) {
	bs, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Write(bs)
}

func (a *apiServer) SearchJob(jobid string) (result interface{}, err error) {

	jobInfo := a.ct.GetJob(jobid)
	if jobInfo == nil {
		result = make(map[string]interface{})
	} else {
		var ok bool
		result, ok = jobInfo.(map[string]interface{})
		if !ok {
			err = errors.New(fmt.Sprintf("jobInfo's type [%T] is not map[string]interface{}", jobInfo))
		}
	}
	return
}


func (a *apiServer) SearchAllJob() (result interface{}, err error) {
	result = a.ct.GetAllJob()
	err = nil
	return
}


func (a *apiServer) ReloadAllJob() (result interface{}, err error) {
	result = a.ct.ReloadAllJobs()
	err = nil
	return
}

func (a *apiServer) DeleteJob(jobid string) (result interface{}, err error) {

	jobInfo, err := a.ct.DelJob(jobid)
	fmt.Println(jobid)
	fmt.Println(jobInfo)
	fmt.Println(err)
	if err != nil {
		return
	}

	if jobInfo == nil {
		result = make(map[string]interface{})
	} else {
		var ok bool
		result, ok = jobInfo.(map[string]interface{})
		if !ok {
			err = errors.New(fmt.Sprintf("jobInfo's type [%T] is not map[string]interface{}", jobInfo))
		}
	}
	return
}


func (a *apiServer) AddJob(jobid string, jobConfigString string) (result interface{}, err error) {

	jobInfo, err := a.ct.AddJob(jobid, jobConfigString)

	if err != nil {
		return
	} else {
		var ok bool
		result, ok = jobInfo.(map[string]interface{})
		if !ok {
			err = errors.New(fmt.Sprintf("jobInfo's type [%T] is not map[string]interface{}", jobInfo))
		}
	}
	return
}


func (a *apiServer) UpdateJob(jobid string, jobConfigString string) (result interface{}, err error) {

	jobInfo, err := a.ct.UpdateJob(jobid, jobConfigString)

	if err != nil {
		return
	} else {
		var ok bool
		result, ok = jobInfo.(map[string]interface{})
		if !ok {
			err = errors.New(fmt.Sprintf("jobInfo's type [%T] is not map[string]interface{}", jobInfo))
		}
	}
	return
}



/*
api 对外提供接口：
	查性能
	任务操作：增加、删除、修改、查询（单个和全部）
	当前实例的信息：实例总数、本实例编号、获得的配置内容

controller:
	控制并发（要知道当前活着的 Goroutine 有哪些，应该有哪些，哪些异常退出了，缺了的能创建，多出来的就关掉）


考虑每个任务放在一个go里面，这样一来就需要检测任务是不是挂了，挂了就要重新启动

任务必须写明是否有周期性
如果有周期性，为了保证周期任务的准确性，需要产生新协程来执行，这样就要求，新协程内的代码必须有超时结束机制
即，周期性的执行体必须自带超时结束机制

数据的返回由执行体自身实施，还是controller统一收集呢？
*/

