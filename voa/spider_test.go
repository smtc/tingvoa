package voa

import (
	"fmt"
	"testing"
	"time"
)

func TestParseTm(t *testing.T) {
	tm, _ := time.Parse("2006-01-02", "2014-09-05")
	fmt.Println(tm)
}

func testDownload(t *testing.T) {
	_, err := downloadMp3("http://stream.51voa.com/201410/se-ws-wildcat-congressmen-money-oil-wells-and-strikers-04oct14.mp3")
	if err != nil {
		fmt.Println(err)
		t.Fatal("download voa file failed")
	}

}
