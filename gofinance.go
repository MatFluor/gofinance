/*
Copyright 2016, Matthias Fluor

This is a simple budgeting web app, based on the blog-post by Alex Recker:
https://alexrecker.com/our-new-sid-meiers-civilization-inspired-budget.html

*/
package main

import (
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func main() {
	// Creates the table on first run if it doesn't exist
	const dbpath = "gofin.db"
	db = initDB(dbpath)
	defer db.Close()
	CreateTable(db)
	// Setting up the routes - handlers in handlers.go
	router := httprouter.New()
	router.GET("/", renderMain)
	router.GET("/stats", handleStats)
	router.GET("/new/transaction", renderInsert)
	router.GET("/new/fixed", renderNewFix)
	router.GET("/edit/:type/:id", handleEdit)
	router.GET("/categories", handleCats)
	router.POST("/confirm/new/transaction", getInput)
	router.POST("/confirm/edit/:type/:id", editEntry)
	router.POST("/confirm/new/fixed", getFixInput)
	router.POST("/confirm/categories", updateCats)
	// Start the Webserver
	err := http.ListenAndServe(":8080", router)
	if err != nil {
		log.Fatal("ListenAndServe: ", router)
	}
}
