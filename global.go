package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
)

func setupLogging(cc *cli.Context) {
	if cc.GlobalBool("verbose") {
		log.SetLevel(log.DebugLevel)
	}
}
