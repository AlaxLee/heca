package heca

import (
	log "github.com/cihub/seelog"
)


func InitLogger() error {
	logger, err := log.LoggerFromConfigAsFile(Home + "/conf/logger.xml")

	if err != nil {
		return err
	}

	log.ReplaceLogger(logger)

	return nil
}
