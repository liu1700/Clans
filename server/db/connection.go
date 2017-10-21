package db

import (
	_ "github.com/go-sql-driver/mysql"
)

var db *DB

func InitDB(ip string, port int, userName string, password string, dbName string) {
	// sql.Open("mysql")
}
