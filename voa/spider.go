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
	voaHost    = "http://tingvoa.me"
)

type VoaItem struct {
	href      string    `sql:"-" json:"-"`
	Id        int64     `json:"id"`
	Typ       string    `sql:"size:64" json:"-"`
	Mp3       string    `sql:"size:256" json:"mp3"`
	Content   string    `sql:"size:60000" json:"content"`
	Title     string    `sql:"size:128" json:"title"`
	Duration  int       `json:"duration"` // seconds
	Image     string    `sql:"size:256" json:"image"`
	Published time.Time `json:"-"`
}

func Voa() error {
	doc, err := goquery.NewDocument(voaSpecial)
	if err != nil {
		return err
	}
	items := []VoaItem{}
	doc.Find("#list > li").Each(func(i int, s *goquery.Selection) {
		hrefs := s.Find("a[href]")
		if hrefs.Length() != 2 {
			return
		}
		a1 := hrefs.First()
		a2 := hrefs.Last()
		href, _ := a2.Attr("href")
		item := VoaItem{
			Typ:   a1.Text(),
			href:  href,
			Title: a2.Text(),
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
func clearItem(item *VoaItem) error {
	item.Typ = strings.Replace(item.Typ, "[", "", -1)
	item.Typ = strings.Replace(item.Typ, "]", "", -1)
	item.Typ = strings.TrimSpace(item.Typ)

	segs := strings.Split(item.Title, "(")
	if len(segs) < 2 {
		return fmt.Errorf("item title format error: %s", item.Title)
	} else if len(segs) == 2 {
		item.Title = strings.TrimSpace(segs[0])
		tm := strings.Replace(segs[1], ")", "", -1)
		item.Published = parseTm(tm)
	} else {
		tm := strings.Replace(segs[len(segs)-1], ")", "", -1)
		item.Title = strings.Join(segs[0:len(segs)-1], "(")
		item.Published = parseTm(tm)
	}

	return nil
}

// 访问网页内容，download mp3文件
// 写入数据库中
func handleItem(item *VoaItem) error {
	doc, err := goquery.NewDocument(item.href)
	if err != nil {
		return err
	}
	if err = handleMp3(item); err != nil {
		return err
	}
	content := ""
	doc.Find("#content > p").Each(func(i int, s *goquery.Selection) {
		content += s.Text()
	})
	item.Content = content
	item.Image, _ = doc.Find("#content > .contentImage > img").First().Attr("src")
	return nil
}

// parse time
//
func parseTm(tm string) time.Time {
	tm = strings.TrimSpace(tm)
	t, _ := time.Parse("2006-01-02", tm)
	return t
}

func handleMp3(item *VoaItem) error {
	fn, err := downloadMp3(item.href)
	if err != nil {
		return err
	}

	item.Mp3 = voaHost + "/assets/voa/" + fn

	return mp3Info(voaAssets+fn, item.Title)
}

// 确认该item所在的目录(yyyy-mm)存在, 如果不存在，创建目录
func mp3Dir(item *VoaItem) (string, error) {

}

// 下载mp3
func downloadMp3(item *VoaItem) (string, error) {
	mp3Name := goutils.ObjectId() + ".mp3"
	mp3File, err := os.Create(voaAssets + mp3Name)
	if err != nil {
		return "", err
	}
	defer mp3File.Close()

	resp, err := http.Get(item.href)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	_, err = io.Copy(mp3File, resp.Body)
	if err != nil {
		return "", err
	}

	return mp3Name, nil
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
