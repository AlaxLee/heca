package main

import (
	"github.com/AlaxLee/heca"
	"net/http"
	"net/url"
	"log"
)

func main() {

	err := heca.InitConfig()
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

		w.Write([]byte("success"))
	})

	log.Fatal(http.ListenAndServe(u.Host, nil))
}