package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
    "os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
)

func continuous() {
	log.WithFields(log.Fields{"func": "continuous"}).Infof("Starting continuous mode with Config: %+v", brf)

	startRestAPI()

	ticker := time.NewTicker(time.Duration(polltime) * time.Second)
	defer ticker.Stop()

	sigC := make(chan os.Signal)
	signal.Notify(sigC, syscall.SIGINT, syscall.SIGTERM)

	backup()
	triggerHealth()

	for {
		select {
		case <-ticker.C:
			backup()
			triggerHealth()
		case <-sigC:
			log.WithFields(log.Fields{"func": "continuous"}).Infof("Signal received, exiting continuous mode")
			return
		}
	}
}

func backup(){

	success := false
	for {
		based = strconv.FormatInt(time.Now().Unix(), 10)
		switch brf.InfraService {
		case INFRA_SERVICE_ZK:
			success = backupZK()
		case INFRA_SERVICE_ETCD:
			success = backupETCD()
		case INFRA_SERVICE_CONSUL:
			success = backupCONSUL()
		default:
			log.WithFields(log.Fields{"func": "continuous"}).Fatal(fmt.Sprintf("Infra service %s unknown or not yet supported", brf.InfraService))
		}
		if !success {
			log.WithFields(log.Fields{"func": "continuous"}).Fatal("Backup was not successfull!")
		}
	}
}
