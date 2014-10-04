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
