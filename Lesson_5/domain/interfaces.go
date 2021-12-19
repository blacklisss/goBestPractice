package domain

import "context"

type Crawler interface {
	Scan(ctx context.Context, url string, curDepth int)
	GetResultChan() <-chan CrawlResult
	Wait()
	Start()
	IcreaseMaxDepth(i int)
	getMaxDepth() int
}

type CrawlResult struct {
	Title string
	Url   string
	Err   error
}
