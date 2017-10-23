package db

import (
	"Clans/server/log"
	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var db *gorm.DB

func InitDB(ip string, port int, userName string, password string, dbName string) {
	// userName:password@tcp(host:port)
	args := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", userName, password, ip, port, dbName)
	var err error
	db, err = gorm.Open("mysql", args)
	if err != nil {
		log.Logger().Error("error when dial db ", err.Error())
		return
	}
}

func CheckConnecting() bool {
	if db != nil {
		if err := db.DB().Ping(); err == nil {
			return true
		} else {
			log.Logger().Error("error when ping %s", err.Error())
		}
	} else {
		log.Logger().Error("db is nil")
	}
	return false
}

func CloseDB() {
	if db != nil {
		if err := db.Close(); err != nil {
			log.Logger().Error("error when close db ", err.Error())
			return
		}
		log.Logger().Debug("db closed")
	}
}

func DB() *gorm.DB {
	return db
}
