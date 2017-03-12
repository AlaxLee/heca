package main

import (
	"github.com/AlaxLee/heca"
)


func main() {

        err := heca.InitLogger()

        if err != nil {
                panic(err)
        }


	c := heca.NewController()

	c.Start()
}
