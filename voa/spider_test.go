package voa

import (
	"github.com/guotie/config"
	"github.com/guotie/deferinit"
	"testing"
)

func init() {
	config.ReadCfg("./config.json")
	deferinit.InitAll()
	createTable()
}

func TestVoa(t *testing.T) {
	err := Voa()
	if err != nil {
		t.Fatal(err)
	}
}
