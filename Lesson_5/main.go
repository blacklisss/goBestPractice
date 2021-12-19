package main

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"l1g2/configuration"
	cr "l1g2/services/crawler"
	"l1g2/services/processor"
	"l1g2/services/requester"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	pid := os.Getpid()
	log.Infof("My PID is: %d\n", pid)

	config, err := configuration.Load("configuration/config.yaml")
	if err != nil {
		log.Fatalf("can't load configuration file")
	}

	var r cr.Requester
	r = requester.NewRequester(time.Minute, config.StartUrl)

	ctx, cancel := context.WithCancel(context.Background())
	crawler := cr.NewCrawler(2, r)
	crawler.Scan(ctx, config.StartUrl, 0)

	chSig := make(chan os.Signal)
	signal.Notify(chSig, syscall.SIGTERM, syscall.SIGINT, syscall.SIGUSR1)

	go processor.ProcessResult(ctx, crawler.GetResultChan(), cancel)
	for {
		select {
		case v := <-chSig:
			if v == syscall.SIGUSR1 {
				crawler.IncreaseMaxDepth(2)
				fmt.Printf("MaxDepth chanched to %d\n", crawler.GetMaxDepth())
			} else {
				log.Infof("Signal SIGTERM catched\n")
				cancel()
			}
		case <-ctx.Done():
			log.Warnf("context canceled\n")
			return
		}
	}
}
