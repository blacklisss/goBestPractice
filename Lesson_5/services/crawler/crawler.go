package crawler

import (
	"context"
	"errors"
	"l1g2/domain"
	"sync"
	"time"
)

type Page interface {
	GetTitle(ctx context.Context) string
	GetLinks() []string
}

type Requester interface {
	GetPage(ctx context.Context, url string) (Page, error)
}

type crawler struct {
	maxDepth  int
	req       Requester
	res       chan domain.CrawlResult
	visited   map[string]struct{}
	visitedMu sync.RWMutex
	page      Page
}

func (c *crawler) GetResultChan() <-chan domain.CrawlResult {
	return c.res
}

func (c *crawler) IncreaseMaxDepth(i int) {
	c.maxDepth += i
}

func (c *crawler) GetMaxDepth() int {
	return c.maxDepth
}

func (c *crawler) Scan(ctx context.Context, url string, curDepth int) {
	var err error
	ctx2, _ := context.WithTimeout(ctx, time.Second*2)
	ctx3, _ := context.WithTimeout(ctx, time.Second*5)

	c.visitedMu.RLock()
	if _, ok := c.visited[url]; ok {
		c.visitedMu.RUnlock()
		return
	}
	c.visitedMu.RUnlock()

	if curDepth >= c.maxDepth {
		//log.Infof("Max Depth on URL: %s\n", url)
		return
	}

	select {
	case <-ctx.Done():
		return
	default:
		c.page, err = c.req.GetPage(ctx3, url)

		c.visitedMu.Lock()
		c.visited[url] = struct{}{}
		c.visitedMu.Unlock()

		if err != nil {
			c.res <- domain.CrawlResult{Url: url, Err: err}
			return
		}

		title := c.page.GetTitle(ctx2)

		if title == "" {
			c.res <- domain.CrawlResult{Url: url, Err: errors.New("page title timeout")}
			return
		}

		c.res <- domain.CrawlResult{
			Title: title,
			Url:   url,
			Err:   nil,
		}

		links := c.page.GetLinks()
		for _, link := range links {
			go c.Scan(ctx, link, curDepth+1)
		}
	}
}

func NewCrawler(maxDepth int, req Requester) *crawler {
	return &crawler{
		maxDepth: maxDepth,
		req:      req,
		res:      make(chan domain.CrawlResult, 100),
		visited:  make(map[string]struct{}),
	}
}
