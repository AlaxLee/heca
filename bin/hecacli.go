package main

import (
	"github.com/AlaxLee/heca"
	"net/url"
	"net/http"
	"io/ioutil"
	"fmt"
	"os"
	"strings"
)


type Command struct {
	Name        string
	Description string
}

var hecaCommands = []Command{
	{"status",    "display job's running status"},
	{"search",    "search job's running status and monitor configs"},
	{"searchall", "search all jobs's running status and monitor configs"},
	{"add",       "add a job to heca"},
	{"del",       "delete a job from heca"},
	{"update",    "update a job to heca"},
	{"reloadall", "reloadall jobs in heca"},
}

var HecaCommands = make(map[string]Command)

func init() {
	for _, cmd := range hecaCommands {
		HecaCommands[cmd.Name] = cmd
	}
}

var ThisCommandName string

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

	s := strings.SplitN(os.Args[0], "/", -1)
	ThisCommandName = s[len(s) - 1]

	c := &cli{}

	if len(os.Args) < 2 {
		c.Usage()
		return
	}

	methodName := os.Args[1]

	if _, ok := HecaCommands[methodName]; !ok {
		c.Usage()
		return
	}

	params := os.Args[2:]

	switch methodName {
	case "status":
		c.status(params)
	case "search":
		if len(params) > 0 {
			c.search(params)

		} else {
			c.Usage()
		}
	case "searchall":
		c.searchall()
	case "del":
		if len(params) > 0 {
			c.del(params[0])

		} else {
			c.Usage()
		}
	case "reloadall":
		c.reloadall()
	case "add":
		if len(params) > 1 {
			c.add(params[0], params[1])

		} else {
			c.Usage()
		}
	case "update":
		if len(params) > 1 {
			c.update(params[0], params[1])

		} else {
			c.Usage()
		}
	}


	//fmt.Printf("%#v\n", params)
	//
	//
	//method := reflect.ValueOf(c).MethodByName(methodName)
	//if method.IsValid() {
	//	return method.Interface().(func(...string) error), nil
	//}



}


type cli struct {
}

func (c *cli) Usage() {
	fmt.Printf("Usage:\n\n\t%s command [arguments]\n\nThe commands are:\n\n", ThisCommandName)
	for _, cmd := range hecaCommands {
		fmt.Printf("\t%-12s%s\n", cmd.Name, cmd.Description)
	}
	fmt.Printf("\nUse \"%s help [command]\" for more information about a command.", ThisCommandName)
}


func (c *cli) status(jobids []string) {
	resp, err := http.PostForm("http://" + heca.Config().Api.ListenAddress + "/api/job/status",
		url.Values{"jobids": jobids})

	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(body))
}


func (c *cli) search(jobids []string) {
	resp, err := http.PostForm("http://" + heca.Config().Api.ListenAddress + "/api/job/search",
		url.Values{"jobids": jobids})

	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(body))
}


func (c *cli) searchall() {
	resp, err := http.Get("http://" + heca.Config().Api.ListenAddress + "/api/job/searchall")

	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(body))
}


func (c *cli) del(jobid string) {
	resp, err := http.PostForm("http://" + heca.Config().Api.ListenAddress + "/api/job/delete",
		url.Values{"jobid": {jobid}})

	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(body))
}


func (c *cli) reloadall() {
	resp, err := http.Get("http://" + heca.Config().Api.ListenAddress + "/api/job/reloadall")

	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(body))
}

//for example:
//	jobid: qidian
//	config: { "endpoint": "qidian", "address": "www.qidian.com",   "jobinterval": 15,   "jobtype": "ping",   "retry": 3,   "timeout": 3 }
func (c *cli) add(jobid string, config string) {
	resp, err := http.PostForm("http://" + heca.Config().Api.ListenAddress + "/api/job/add",
		url.Values{"jobid": {jobid}, "config": {config}})

	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(body))
}



func (c *cli) update(jobid string, config string) {
	resp, err := http.PostForm("http://" + heca.Config().Api.ListenAddress + "/api/job/update",
		url.Values{"jobid": {jobid}, "config": {config}})

	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(body))
}