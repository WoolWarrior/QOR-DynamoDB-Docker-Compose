package main

import (
	"qor-started/admin"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
)

func main() {

	// Set up an unused virtual database in memory
	DB, _ := gorm.Open("sqlite3", ":memory:")

	// Set up web object for Gin framework
	r := gin.New()

	// Set up QOR admin object
	a := admin.New(DB, "", "secret")

	// Bind QOR admin to Gin
	a.Bind(r)

	// Run web app on given address and port
	r.Run("0.0.0.0:8080")
}
