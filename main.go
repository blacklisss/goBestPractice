package main

// File is not `goimports`-ed with -local github.com/golangci/golangci-lint (goimports)
import (
	"context"
	"fmt"
	"gb/bestPractice/configuration"
	cr "gb/bestPractice/services/crawler"
	"gb/bestPractice/services/processor"
	"gb/bestPractice/services/requester"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

func main() {
	pid := os.Getpid()
	log.Infof("My PID is: %d\n", pid)

	config, err := configuration.Load("configuration/config.yaml")
	if err != nil {
		log.Fatalf("can't load configuration file")
	}
	// S1021: should merge variable declaration with assignment on next line (gosimple)
	var r cr.Requester = requester.NewRequester(time.Minute, config.StartURL)

	ctx, cancel := context.WithCancel(context.Background())
	crawler := cr.NewCrawler(2, r)
	crawler.Scan(ctx, config.StartURL, 0)

	chSig := make(chan os.Signal, 1) // sigchanyzer: misuse of unbuffered os.Signal channel as argument to signal.Notify (govet)
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
