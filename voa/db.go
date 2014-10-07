package voa

import (
	"github.com/smtc/justcms/database"
)

func createTable() {
	db := database.GetDB("")
	db.AutoMigrate(VoaItem{})
}

func saveItem(item *VoaItem) error {
	db := database.GetDB("")
	return db.Save(item).Error
}

// 获得数据库中最新的id
func lastItemId() int64 {
	var item VoaItem
	db := database.GetDB("")
	if err := db.Order("orig_id desc").Find(&item).Limit(1).Error; err != nil {
		return 0
	}
	return item.OrigId
}
