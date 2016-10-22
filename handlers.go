/*
This file holds all the handlers - to keep the main go file a bit cleaner
*/
package main

import (
	"database/sql"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
)

// Make the DB global for all
var db *sql.DB

func handleStatsDetails(w http.ResponseWriter, r *http.Request, pr httprouter.Params) {
	t, _ := template.ParseFiles("templates/details.html", "templates/header.html")
	data := SumByCats(db, pr.ByName("type"))
	t.ExecuteTemplate(w, "details", map[string]interface{}{"data": data, "type": pr.ByName("type"), "mapping": false})
}

func handleSummaryDetails(w http.ResponseWriter, r *http.Request, pr httprouter.Params) {
	t, _ := template.ParseFiles("templates/details.html", "templates/header.html")
	data := SumSummary(db, pr.ByName("type"))
	t.ExecuteTemplate(w, "details", map[string]interface{}{"data": data, "type": pr.ByName("type"), "mapping": true})
}

func handleCats(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	t, _ := template.ParseFiles("templates/editcategories.html", "templates/header.html")
	items := getCategories(db)
	t.ExecuteTemplate(w, "categories", items)
}

func updateCats(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	r.ParseForm()
	var cats []Category
	for key, values := range r.Form { // range over map
		keylist := strings.Split(key, "_")
		for _, value := range values { // range over []string
			var id int
			if keylist[0] != "" {
				id, _ = strconv.Atoi(keylist[0])
			} else {
				id = 0
			}
			nullID := ToNullInt64(id)
			cats = append(cats, Category{ID: nullID, Description: keylist[1], Mapping: ToNullString(value)})
		}
	}
	UpdateCats(db, cats)
	http.Redirect(w, r, "/", 301)
}

func handleStats(w http.ResponseWriter, r *http.Request, pr httprouter.Params) {
	t, _ := template.ParseFiles("templates/stats.html", "templates/header.html")
	// Get labels and values for stats concurrently
	daylabelchan := make(chan []string)
	typelabelchan := make(chan []string)
	monlabelchan := make(chan []string)
	dayvalchan := make(chan []float64)
	typevalchan := make(chan []float64)
	monvalchan := make(chan []float64)
	go sumUp(db, "daily", daylabelchan, dayvalchan)
	go sumUp(db, "type", typelabelchan, typevalchan)
	go sumUp(db, "monthly", monlabelchan, monvalchan)
	dayLabels := <-daylabelchan
	dayValues := <-dayvalchan
	typeLabels := <-typelabelchan
	typeValues := <-typevalchan
	monLabels := <-monlabelchan
	monValues := <-monvalchan
	// Calculate the correct numbers by day
	magicNumber := baseMagic(db)
	for i := 0; i < len(dayValues); i++ {
		dayValues[i] = magicNumber - (dayValues[i] * -1)
	}
	// Slice of slices for the table
	type category struct {
		Descr   string
		Val     float64
		Percent float64
	}
	// Calculate the percentage for each category
	var totalamount float64
	for _, val := range typeValues {
		if val < 0 {
			val *= -1
		}
		totalamount += val
	}
	// Get the categories and sum them up to display in the table
	var catList []category
	for i := 0; i < len(typeLabels); i++ {
		percentage := percentages(totalamount, typeValues[i])
		catList = append(catList, category{Descr: typeLabels[i], Val: typeValues[i], Percent: percentage})
	}
	for i, val := range monLabels {
		monInt, _ := strconv.Atoi(val)
		numdays := daysInMonth(time.Now().Year(), time.Month(monInt))
		monValues[i] = monValues[i] + float64(numdays)*magicNumber
	}
	t.ExecuteTemplate(w, "stats", map[string]interface{}{"dayLabels": dayLabels, "dayValues": dayValues,
		"magicnumber": magicNumber, "types": catList, "monLabels": monLabels, "monValues": monValues})
}

func handleEdit(w http.ResponseWriter, r *http.Request, pr httprouter.Params) {
	t, _ := template.ParseFiles("templates/edit.html", "templates/header.html")
	entryID := pr.ByName("id")
	var entry int
	entry, _ = strconv.Atoi(entryID)
	trans := getSingle(db, entry, pr.ByName("type"))
	if trans.Amount < 0 {
		trans.Amount = trans.Amount * -1
	}
	var fixcheck bool
	if pr.ByName("type") == "fixed" {
		fixcheck = true
	}
	t.ExecuteTemplate(w, "edit", map[string]interface{}{"trans": trans, "transtype": pr.ByName("type"), "fixcheck": fixcheck})
}

func editEntry(w http.ResponseWriter, r *http.Request, pr httprouter.Params) {
	r.ParseForm()
	income := false
	description := r.Form["description"][0]
	amountstr := r.Form["amount"][0]
	amount, erra := strconv.ParseFloat(amountstr, 64)
	if erra != nil {
		panic(erra)
	}
	incomecheck := r.Form["income"]
	if len(incomecheck) == 0 {
		income = false
	} else {
		income = true
	}
	recurrence := ""
	if pr.ByName("type") == "fixed" {
		recurrence = strings.ToLower(r.Form["recurrence"][0])
	}
	idstr := pr.ByName("id")
	idint, _ := strconv.Atoi(idstr)
	ChangeItem(db, Transaction{ID: idint, Description: description, Amount: amount, Income: income, Recurrence: recurrence}, pr.ByName("type"))
	// Get back to the main page
	http.Redirect(w, r, "/", 301)
}

func getInput(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	r.ParseForm()
	income := false
	description := r.Form["description"][0]
	amountstr := r.Form["amount"][0]
	amount, erra := strconv.ParseFloat(amountstr, 64)
	if erra != nil {
		panic(erra)
	}
	incomecheck := r.Form["income"]
	if len(incomecheck) == 0 {
		income = false
	} else {
		income = true
	}
	StoreItem(db, Transaction{Description: description, Amount: amount, Income: income}, "transaction")
	// Get back to the main page
	http.Redirect(w, r, "/", 301)
}

func getFixInput(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	r.ParseForm()
	income := false
	description := r.Form["description"][0]
	amountstr := r.Form["amount"][0]
	recurrence := r.Form["recurrence"][0]
	amount, erra := strconv.ParseFloat(amountstr, 64)
	if erra != nil {
		panic(erra)
	}
	incomecheck := r.Form["income"]
	if len(incomecheck) == 0 {
		income = false
	} else {
		income = true
	}
	influence := calcRate(Transaction{Recurrence: recurrence, Amount: amount, Income: income})
	StoreItem(db, Transaction{Description: description, Amount: amount, Income: income, Recurrence: recurrence, Influence: influence}, "fixed")
	// Get back to the main page
	http.Redirect(w, r, "/", 301)
}

// Handler to display the main page - with db-values
func renderMain(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	t, err := template.ParseFiles("templates/index.html", "templates/header.html")
	if err != nil {
		panic(err)
	}
	// Read the Database to get the current stuff (Date = today)
	fixed := ReadItem(db, "fixed")
	trans := ReadItem(db, "transaction")
	magicNumber := baseMagic(db)
	currentNumber := currentMagic(db)
	weekchan := make(chan float64)
	monthchan := make(chan float64)
	yearchan := make(chan float64)
	go expensesPerPeriod("week", weekchan)
	go expensesPerPeriod("month", monthchan)
	go expensesPerPeriod("year", yearchan)
	weektotal := <-weekchan
	monthtotal := <-monthchan
	yeartotal := <-yearchan
	t.ExecuteTemplate(w, "index", map[string]interface{}{"fix": fixed, "tran": trans,
		"mn": magicNumber, "curr": currentNumber,
		"weektotal": weektotal, "monthtotal": monthtotal, "yeartotal": yeartotal})
}

// Handler for the insertion
func renderInsert(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	t, err := template.ParseFiles("templates/input.html", "templates/header.html")
	if err != nil {
		panic(err)
	}
	t.ExecuteTemplate(w, "input", "")
}

// Handler for the insertion
func renderNewFix(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	t, err := template.ParseFiles("templates/inputfix.html", "templates/header.html")
	if err != nil {
		panic(err)
	}
	t.ExecuteTemplate(w, "inputfix", "")
}
