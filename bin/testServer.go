package main

import (
	"github.com/AlaxLee/heca"
	"net/http"
	"net/url"
	log "github.com/cihub/seelog"
	"io/ioutil"
	"encoding/json"
)

func main() {

	var err error
	err = heca.InitLogger()
	if err != nil {
		panic(err)
	}

	err = heca.InitConfig()
	if err != nil {
		panic(err)
	}

	apiUrl := heca.Config().Push.APIUrl

	u, err := url.Parse(apiUrl)
	if err != nil {
		panic(err)
	}

	http.HandleFunc(u.Path, func(w http.ResponseWriter, req *http.Request) {
		if req.ContentLength == 0 {
			http.Error(w, "body is blank", http.StatusBadRequest)
			return
		}

		reqBody, err := ioutil.ReadAll(req.Body)
		if err != nil {
			http.Error(w, "read body failed: " + err.Error(), http.StatusBadRequest)
			return
		}

		b := make([]map[string]interface{}, 0)
		err = json.Unmarshal(reqBody, &b)
		if err != nil {
			http.Error(w, "body is not a array json: " + err.Error(), http.StatusBadRequest)
			return
		}

		log.Infof("get %d metrics", len(b))

		w.Write([]byte("success"))
	})

	log.Error(http.ListenAndServe(u.Host, nil))
}