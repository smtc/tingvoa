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
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	_          = io.Copy
	_          = http.Get
	voaSpecial = "http://www.51voa.com/VOA_Special_English/"
	voaAssets  = "../app/assets/voa/"
	voaHost    = "http://tingvoa.me"
	host51voa  = "http://www.51voa.com"
	voaIdRe    = regexp.MustCompile(`^\d+$`)
)

type VoaItem struct {
	href         string    `sql:"-" json:"-"`
	downloadHref string    `sql:"-" json:"-"`
	lyricHref    string    `sql:"-" json:"-"`
	Id           int64     `json:"id"`
	OrigId       int64     `json:"-" sql:"not null;unique"`
	Typ          string    `sql:"size:64" json:"-"`
	Mp3          string    `sql:"size:256" json:"mp3"`
	Content      string    `sql:"type:TEXT;" json:"content"`
	Title        string    `sql:"size:128" json:"title"`
	Duration     int       `json:"duration"` // seconds
	Image        string    `sql:"size:256" json:"image"`
	Lyric        string    `sql:"type:TEXT;" json:"lyric"`
	Published    time.Time `json:"-"`
}

func Voa() error {
	doc, err := goquery.NewDocument(voaSpecial)
	if err != nil {
		log.Printf("Get Url %s failed: %v\n", voaSpecial, err)
		return err
	}
	lastId := lastItemId()

	items := []VoaItem{}
	doc.Find("#list li").Each(func(i int, s *goquery.Selection) {
		hrefs := s.Find("a[href]")
		if hrefs.Length() < 2 {
			log.Printf("element %d no more than 2 href\n", i)
			return
		}
		a1 := hrefs.First()
		a2 := hrefs.Last()
		href, _ := a2.Attr("href")
		origId := voaId(href)
		if origId == 0 {
			log.Printf("url(%d %s) has no id or format changed.\n", i, href)
			return
		}
		if !(strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://")) {
			href = host51voa + href
		}

		// 该item应该已经获取过
		if origId <= lastId {
			return
		}
		item := VoaItem{
			Typ:    a1.Text(),
			href:   href,
			OrigId: origId,
			Title:  a2.Text(),
		}
		if err := clearItem(&item); err == nil {
			items = append(items, item)
		} else {
			log.Printf("clear item %v failed: %v\n", item, err)
		}
	})

	if len(items) == 0 {
		return fmt.Errorf("Not found voa special english item.")
	}

	// download mp3, lrc, get content
	for _, item := range items {
		if err := handleItem(&item); err == nil {
			if err = saveItem(&item); err != nil {
				log.Printf("save item %s failed: %v\n", item.href, err)
			}
		} else {
			log.Printf("handle item %s failed: %v\n", item.href, err)
		}
	}

	return nil
}

// http://www.51voa.com/VOA_Special_English/surveillance-software-key-concern-at-internet-governance-meeting-58809.html
//
func voaId(href string) int64 {
	segs := strings.Split(href, "-")
	if len(segs) <= 1 {
		segs = strings.Split(href, "_")
		if len(segs) <= 1 {
			return 0
		}
	}
	idhtml := strings.Split(segs[len(segs)-1], ".")
	if len(idhtml) <= 1 {
		return 0
	}
	id := idhtml[0]
	if voaIdRe.MatchString(id) {
		oid, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			return 0
		}
		return oid
	}
	return 0
}

// typ: remove [,], trim space, ex: [ Education Report ]
// title: ex: Is a College Education Worth the Price?  (2014-10-4)
func clearItem(item *VoaItem) error {
	var (
		tm  string
		err error
	)

	item.Typ = strings.Replace(item.Typ, "[", "", -1)
	item.Typ = strings.Replace(item.Typ, "]", "", -1)
	item.Typ = strings.TrimSpace(item.Typ)

	segs := strings.Split(item.Title, "(")
	if len(segs) < 2 {
		return fmt.Errorf("item title format error: %s", item.Title)
	} else if len(segs) == 2 {
		item.Title = strings.TrimSpace(segs[0])
		tm = strings.Replace(segs[1], ")", "", -1)
	} else {
		tm = strings.Replace(segs[len(segs)-1], ")", "", -1)
		item.Title = strings.Join(segs[0:len(segs)-1], "(")
	}

	item.Published, err = time.Parse("2006-1-2", strings.TrimSpace(tm))
	if err != nil {
		return err
	}

	return nil
}

func addHostPrefix(href string) string {
	if href == "" {
		return href
	}
	if !(strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://")) {
		return host51voa + href
	}
	return href
}

// 访问网页内容，download mp3文件
// 写入数据库中
func handleItem(item *VoaItem) error {
	doc, err := goquery.NewDocument(item.href)
	if err != nil {
		return err
	}
	var exist bool
	if item.downloadHref, exist = doc.Find("#mp3").Attr("href"); !exist {
		return fmt.Errorf("mp3 download href not exist")
	}
	item.downloadHref = addHostPrefix(item.downloadHref)

	item.lyricHref, _ = doc.Find("#lrc").Attr("href")
	item.lyricHref = addHostPrefix(item.lyricHref)

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

func handleMp3(item *VoaItem) error {
	fn, err := downloadMp3(item)
	if err != nil {
		return err
	}

	if _, err := downloadLyric(item); err != nil {
		log.Printf("download lyric for %d failed: %v\n", item.OrigId, err)
	}

	yyyymm := fmt.Sprintf("%04d%02d", item.Published.Year(), item.Published.Month())

	item.Mp3 = voaHost + "/assets/voa/" + yyyymm + "/" + fn

	return mp3Info(voaAssets+yyyymm+"/"+fn, item.Title)
}

// 确认该item所在的目录(yyyy-mm)存在, 如果不存在，创建目录
func mp3Dir(item *VoaItem) (string, error) {
	yyyymm := fmt.Sprintf("%04d%02d", item.Published.Year(), item.Published.Month())
	dir := voaAssets + yyyymm
	if err := goutils.CreateDirIfNotExist(dir); err != nil {
		return dir, err
	}
	return dir, nil
}

// 下载mp3
func downloadMp3(item *VoaItem) (string, error) {
	mp3Name := fmt.Sprint(item.OrigId) + ".mp3"
	dir, err := mp3Dir(item)
	if err != nil {
		return "", err
	}
	mp3File, err := os.Create(path.Join(dir, mp3Name))
	if err != nil {
		return "", err
	}
	defer mp3File.Close()

	resp, err := http.Get(item.downloadHref)
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

// 下载歌词
func downloadLyric(item *VoaItem) (string, error) {
	// lyric
	if item.lyricHref == "" {
		return "", nil
	}

	lycName := fmt.Sprint(item.OrigId) + ".lrc"
	dir, err := mp3Dir(item)
	if err != nil {
		return "", err
	}
	lycFile, err := os.Create(path.Join(dir, lycName))
	if err != nil {
		return "", err
	}
	defer lycFile.Close()

	resp, err := http.Get(item.lyricHref)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	_, err = io.Copy(lycFile, resp.Body)
	if err != nil {
		return "", err
	}

	return lycName, nil
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
