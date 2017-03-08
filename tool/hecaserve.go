package main

import (
	"github.com/AlaxLee/heca"
	log "github.com/cihub/seelog"
)


func main() {
	logger, err := log.LoggerFromConfigAsString(`
	<seelog minlevel="info">
		<outputs>
			<console />
		</outputs>
	</seelog>
	`)

	if err != nil {
		panic(err)
	}

	log.ReplaceLogger(logger)


	c := heca.NewController()

	c.Start()
}