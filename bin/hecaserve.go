package main

import (
	"github.com/AlaxLee/heca"
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


	c := heca.NewController()

	c.Start()
}
