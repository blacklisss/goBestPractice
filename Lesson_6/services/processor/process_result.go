package processor

//  File is not `goimports`-ed with -local github.com/golangci/golangci-lint (goimports)
import (
	"context"
	"l1g2/domain"
	"time"

	log "github.com/sirupsen/logrus"
)

func ProcessResult(ctx context.Context, in <-chan domain.CrawlResult, cancel context.CancelFunc) {
	var errCount int
	var ticker = time.NewTicker(time.Second * 1)
	var sleeping time.Time

	for {
		select {
		case res := <-in:
			sleeping = time.Now()
			if res.Err != nil {
				log.Errorf("ERROR Link: %s, err: %v\n", res.URL, res.Err)
				errCount++
				/*if errCount >= 1 { empty branch (staticcheck)
					//cancel()  commentFormatting: put a space between `//` and comment text (gocritic)
				}*/
			} else {
				log.Infof("Link: %s, Title: %s\n", res.URL, res.Title)
			}
		case <-ctx.Done():
			log.Warnf("context canceled\n")
			return
		case <-ticker.C:
			now := time.Now()
			if now.Sub(sleeping).Seconds() > 7 {
				log.Warnf("process timeout\n")
				cancel()
				return
			}
		}
	}
}
