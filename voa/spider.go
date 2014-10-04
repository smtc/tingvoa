package voa

// 获取voa special english
// http://www.51voa.com/VOA_Special_English/
//
// 0 type
// 1 MP3文件
// 2 同步字幕
// 3 英语原文
// 4 插图

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"time"
)

var (
	voaSpecial = "http://www.51voa.com/VOA_Special_English/"
)

type voaItem struct {
	typ   string
	href  string
	title string
	image string
	pub   time.Time
}

func Voa() error {
	doc, err := goquery.NewDocument(voaSpecial)
	if err != nil {
		return err
	}
	items := []voaItem{}
	doc.Find("#list > li").Each(func(i int, s *goquery.Selection) {
		hrefs := s.Find("a[href]")
		if hrefs.Length() != 2 {
			return
		}
		a1 := hrefs.First()
		a2 := hrefs.Last()
		href, _ := a2.Attr("href")
		item := voaItem{
			typ:   a1.Text(),
			href:  href,
			title: a2.Text(),
		}
		items = append(items, item)
	})

	if len(items) == 0 {
		return fmt.Errorf("Not found voa special english item.")
	}

	return nil
}
