package voa

import (
	"github.com/smtc/goutils"
	"github.com/smtc/justcms/database"
)

func createTable() {
	db := database.GetDB("")
	db.AutoMigrate(VoaItem{})
}

func saveItem(item *VoaItem) error {
	return nil
}
