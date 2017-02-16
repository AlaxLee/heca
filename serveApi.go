package heca

import (
	"fmt"
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

