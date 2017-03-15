package heca

import (
	"net/http"
	_ "net/http/pprof"
	"encoding/json"
	log "github.com/cihub/seelog"
)

type apiServer struct {
	ct *Controller
}

func NewApiServer(controller *Controller) *apiServer {
	return &apiServer{ct: controller}
}

func (a *apiServer) start() {

	http.HandleFunc("/api/job/status", func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		result := make(map[string]interface{})

		jobids, ok := r.Form["jobids"]

		if !ok  ||  len(jobids) == 0 {
			result = a.ct.GetAllStatus()
		} else {
			result = a.ct.GetStatus(jobids)
		}

		RenderJson(w, result)
	})

	http.HandleFunc("/api/job/search", func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		jobids, ok := r.Form["jobids"]
		if !ok {
			http.Error(w, "need params jobids", http.StatusInternalServerError)
			return
		}

		result := a.ct.GetJob(jobids)

		RenderJson(w, result)
	})


	http.HandleFunc("/api/job/searchall", func(w http.ResponseWriter, r *http.Request) {

		result := a.ct.GetAllJob()

		RenderJson(w, result)
	})


	http.HandleFunc("/api/job/delete", func(w http.ResponseWriter, r *http.Request) {
		jobid := r.FormValue("jobid")

		result, err := a.ct.DelJob(jobid)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		RenderJson(w, result)
	})

	http.HandleFunc("/api/job/add", func(w http.ResponseWriter, r *http.Request) {
		jobid := r.FormValue("jobid")
		jobConfigString := r.FormValue("config")

		result, err := a.ct.AddJob(jobid, jobConfigString)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		RenderJson(w, result)
	})

	http.HandleFunc("/api/job/update", func(w http.ResponseWriter, r *http.Request) {
		jobid := r.FormValue("jobid")
		jobConfigString := r.FormValue("config")

		result, err := a.ct.UpdateJob(jobid, jobConfigString)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		RenderJson(w, result)
	})


	http.HandleFunc("/api/job/reloadall", func(w http.ResponseWriter, r *http.Request) {

		result := a.ct.ReloadAllJobs()

		RenderJson(w, result)
	})


	go func() {
		log.Critical(http.ListenAndServe(Config().Api.ListenAddress, nil))
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

