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
	id3 "github.com/mikkyang/id3-go"
	"github.com/smtc/goutils"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	voaSpecial = "http://www.51voa.com/VOA_Special_English/"
	voaAssets  = "../app/assets/voa/"
)

type voaItem struct {
	typ      string
	href     string
	title    string
	content  string
	image    string
	duration int // seconds
	pub      time.Time
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
		if err := clearItem(&item); err == nil {
			items = append(items, item)
		}
	})

	if len(items) == 0 {
		return fmt.Errorf("Not found voa special english item.")
	}

	// download mp3, lrc, get content

	return nil
}

// typ: remove [,], trim space, ex: [ Education Report ]
// title: ex: Is a College Education Worth the Price?  (2014-10-4)
func clearItem(item *voaItem) error {
	item.typ = strings.Replace(item.typ, "[", "", -1)
	item.typ = strings.Replace(item.typ, "]", "", -1)
	item.typ = strings.TrimSpace(item.typ)

	segs := strings.Split(item.title, "(")
	if len(segs) < 2 {
		return fmt.Errorf("item title format error: %s", item.title)
	} else if len(segs) == 2 {
		item.title = strings.TrimSpace(segs[0])
		tm := strings.Replace(segs[1], ")", "", -1)
		item.pub = parseTm(tm)
	} else {
		tm := strings.Replace(segs[len(segs)-1], ")", "", -1)
		item.title = strings.Join(segs[0:len(segs)-1], "(")
		item.pub = parseTm(tm)
	}

	return nil
}

// 访问网页内容，download mp3文件
// 写入数据库中
func handleItem() {

}

// parse time
//
func parseTm(tm string) time.Time {
	tm = strings.TrimSpace(tm)
	t, _ := time.Parse("2006-01-02", tm)
	return t
}

func downloadMp3(url string) (string, error) {
	mp3Name := goutils.ObjectId() + ".mp3"
	mp3File, err := os.Create(voaAssets + mp3Name)
	if err != nil {
		return "", err
	}
	defer mp3File.Close()

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	_, err = io.Copy(mp3File, resp.Body)
	if err != nil {
		return "", err
	}

	return voaAssets + mp3Name, nil
}

func mp3Info(fn, title string) error {
	info, err := id3.Open(fn)
	if err != nil {
		return err
	}
	defer info.Close()

	info.SetTitle(title)
	info.SetArtist("ting voa")
	return nil
}
