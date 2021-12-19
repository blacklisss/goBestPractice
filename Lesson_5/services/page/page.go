package page

import (
	"context"
	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"
	"io"
	urlParser "net/url"
	"strings"
)

type page struct {
	doc      *goquery.Document
	StartUrl string
}

func (p page) GetTitle(ctx context.Context) string {
	select {
	case <-ctx.Done():
		return ""
	default:
		return p.doc.Find("title").First().Text()
	}
}

func (p page) GetLinks() []string {
	var urls []string

	startUrlInfo, _ := urlParser.Parse(p.StartUrl)

	p.doc.Find("a").Each(func(_ int, s *goquery.Selection) {
		url, ok := s.Attr("href")
		if ok {
			if strings.HasPrefix(url, "#") {
				//log.Infof("Anchor link %s\n", url)
				return
			}

			urlInfo, err := urlParser.Parse(url)
			if err != nil {
				log.Warnf("error parse URL: %s\n", url)
				return
			}

			if urlInfo.Host == "" {
				if strings.HasPrefix(url, "/") {
					url = startUrlInfo.Scheme + "://" + startUrlInfo.Host + url
				} else {
					url = p.StartUrl + url
				}
			}
			//Здесь может быть относительная ссылка, нужно абсолютную
			urls = append(urls, url)
		}
	})
	return urls
}

func NewPage(raw io.Reader, startUrl string) (page, error) {
	doc, err := goquery.NewDocumentFromReader(raw)
	if err != nil {
		return page{}, err
	}

	return page{doc, startUrl}, nil
}
