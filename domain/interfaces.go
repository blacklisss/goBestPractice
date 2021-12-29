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
	URL   string // ST1003: struct field Url should be URL (stylecheck)
	Err   error
}
