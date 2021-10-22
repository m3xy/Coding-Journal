package main

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

const (
	host     = "127.0.0.1"
	port     = 3600
	user     = "mysql"
	protocol = "tcp"
	password = "secret"
	dbname   = "my-db"
)

func dbInit() error {
	var err error

	// Set MySQL info in DSN format according to Go MySQL Drive -
	// user:password@protocol(host:port)/dbname?[param1=val...]
	mysqlInfo := fmt.Sprintf("%s:%s@%s(%s:%d)/%s", user, password, protocol, host, port, dbname)
	db, err = sql.Open("mariadb", mysqlInfo)
	if err != nil {
		return err
	}

	// Set connection sanity options for database.
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	return nil
}
