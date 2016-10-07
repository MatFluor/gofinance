/*
This file holds all the database-relevant functions
like creation, insertion and reading
*/
package main

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Transaction Basic type
type Transaction struct {
	ID          int
	Amount      float64
	Description string
	Income      bool
	Recurrence  string
	Influence   float64
	Timestamp   time.Time
}

// Category basic struct
type Category struct {
	ID          sql.NullInt64
	Mapping     sql.NullString
	Description string
}

// ToNullInt64 helper to convert from regular int
func ToNullInt64(i int) sql.NullInt64 {
	newI := int64(i)
	return sql.NullInt64{Int64: newI, Valid: true}
}

// FromNullInt64 helper to convert that back
func FromNullInt64(i sql.NullInt64) int64 {
	newI := i.Int64
	return newI
}

// ToNullString helper to convert into Nullstring
func ToNullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func initDB(filepath string) *sql.DB {
	db, err := sql.Open("sqlite3", filepath)
	if err != nil {
		panic(err)
	}
	if db == nil {
		panic("db nil")
	}
	return db
}

// CreateTable to intialize the database
func CreateTable(db *sql.DB) {
	// Create if not exists
	sqlTable := `
  CREATE TABLE IF NOT EXISTS fixed(
    id INTEGER NOT NULL PRIMARY KEY,
    description TEXT,
    amount REAL,
		income BOOL,
		recurrence TEXT,
		influence REAL,
    timestamp DATETIME
    );
    `
	_, err := db.Exec(sqlTable)
	if err != nil {
		panic(err)
	}
	sqlTable2 := `
  CREATE TABLE IF NOT EXISTS transactions(
    id INTEGER NOT NULL PRIMARY KEY,
    description TEXT,
    amount REAL,
		income BOOL,
		recurrence TEXT,
    timestamp DATETIME
    );
    `
	_, err2 := db.Exec(sqlTable2)
	if err2 != nil {
		panic(err2)
	}
	sqlTable3 := `
  CREATE TABLE IF NOT EXISTS mappings(
    id INTEGER NOT NULL PRIMARY KEY,
    mapping TEXT,
    description TEXT
    );
    `
	_, err3 := db.Exec(sqlTable3)
	if err3 != nil {
		panic(err3)
	}
}

// UpdateCats Insert or Replace the categories
func UpdateCats(db *sql.DB, cats []Category) {
	var newID int64
	tx, _ := db.Begin()
	sqlUpdate := "INSERT OR REPLACE INTO mappings (id, mapping, description)	VALUES (?, ?, ?)"
	stmt, _ := tx.Prepare(sqlUpdate)
	for i := 0; i < len(cats); i++ {
		if FromNullInt64(cats[i].ID) == 0 {
			rows := db.QueryRow("SELECT MAX(id) FROM mappings")
			_ = rows.Scan(&newID)
			newID++
		} else {
			newID = FromNullInt64(cats[i].ID)
		}
		_, err2 := stmt.Exec(newID, cats[i].Mapping, cats[i].Description)
		if err2 != nil {
			panic(err2)
		}
	}
	tx.Commit()
}

// StoreItem holds logic to insert a transaction
func StoreItem(db *sql.DB, item Transaction, transtype string) {
	switch transtype {
	case "fixed":
		sqlAddItem := `
		INSERT INTO fixed(
			description,
			amount,
			income,
			recurrence,
			influence,
			timestamp
			) VALUES(?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
			`
		stmt, err := db.Prepare(sqlAddItem)
		if err != nil {
			panic(err)
		}
		defer stmt.Close()
		_, err2 := stmt.Exec(item.Description, item.Amount, item.Income, item.Recurrence, item.Influence)
		if err2 != nil {
			panic(err2)
		}
	case "transaction":
		sqlAddItem := `
	INSERT INTO transactions(
		description,
		amount,
		income,
		timestamp
	) VALUES(?, ?, ?, CURRENT_TIMESTAMP)
	`
		stmt, err := db.Prepare(sqlAddItem)
		if err != nil {
			panic(err)
		}
		defer stmt.Close()
		if item.Income != true {
			item.Amount = -item.Amount
		}
		_, err2 := stmt.Exec(item.Description, item.Amount, item.Income)
		if err2 != nil {
			panic(err2)
		}
	}
}

// ChangeItem holds logic to insert a transaction
func ChangeItem(db *sql.DB, item Transaction, transtype string) {
	switch transtype {
	case "fixed":
		sqlAddItem := `
		UPDATE fixed SET
			description = ?,
			amount = ?,
			income = ?,
			recurrence = ?,
			influence = ?
			WHERE id = ?
			`
		stmt, err := db.Prepare(sqlAddItem)
		if err != nil {
			panic(err)
		}
		defer stmt.Close()
		influence := calcRate(item)
		_, err2 := stmt.Exec(item.Description, item.Amount, item.Income, item.Recurrence, influence, item.ID)
		if err2 != nil {
			panic(err2)
		}
	case "transactions":
		sqlAddItem := `
	UPDATE transactions SET
		description = ?,
		amount = ?,
		income = ?
	WHERE id = ?
	`
		stmt, err := db.Prepare(sqlAddItem)
		if err != nil {
			panic(err)
		}
		defer stmt.Close()
		if item.Income != true {
			item.Amount = -item.Amount
		}
		_, err2 := stmt.Exec(item.Description, item.Amount, item.Income, item.ID)
		if err2 != nil {
			panic(err2)
		}
	}
}

// ReadItem returns the items
func ReadItem(db *sql.DB, transtype string) []Transaction {
	var result []Transaction
	switch transtype {
	case "fixed":
		sqlReadFix := `
		SELECT id, description, amount, income, influence, recurrence FROM fixed
		ORDER BY amount DESC
		`

		rows, err := db.Query(sqlReadFix)
		if err != nil {
			panic(err)
		}
		for rows.Next() {
			item := Transaction{}
			_ = rows.Scan(&item.ID, &item.Description, &item.Amount, &item.Income, &item.Influence, &item.Recurrence)
			result = append(result, item)
		}
	case "transaction":
		sqlReadTrans := `
		SELECT id, description, amount, income FROM transactions
		WHERE datetime(timestamp) >= DATE('now')
		ORDER BY datetime(timestamp) DESC
		`
		rows, err := db.Query(sqlReadTrans)
		if err != nil {
			panic(err)
		}
		for rows.Next() {
			item := Transaction{}
			_ = rows.Scan(&item.ID, &item.Description, &item.Amount, &item.Income)
			result = append(result, item)
		}
	}
	return result
}

func getSingle(db *sql.DB, id int, transtype string) Transaction {
	row := db.QueryRow("SELECT id, description, amount, income, recurrence FROM "+transtype+" WHERE id = ?", id)
	var item Transaction
	_ = row.Scan(&item.ID, &item.Description, &item.Amount, &item.Income, &item.Recurrence)
	return item
}

func baseMagic(db *sql.DB) float64 {
	var magicNumber float64
	sqlRead := `SELECT SUM(influence) FROM fixed`
	row := db.QueryRow(sqlRead)
	_ = row.Scan(&magicNumber)
	return magicNumber
}

// This returns the total of all expenses from a specified period (week, month or year)
func totalExpenses(db *sql.DB, period string) float64 {
	var sqlRead string
	var totalExpenses float64
	switch period {
	case "week":
		sqlRead = "SELECT SUM(amount) FROM transactions WHERE timestamp >= date('now', 'weekday 1', '-7 days');"
	case "month":
		sqlRead = "SELECT SUM(amount) FROM transactions WHERE timestamp >= date('now', 'start of month');"
	case "year":
		sqlRead = "SELECT SUM(amount) FROM transactions WHERE timestamp >= date('now', 'start of year');"
	}
	row := db.QueryRow(sqlRead)
	_ = row.Scan(&totalExpenses)
	return totalExpenses
}

func sumUp(db *sql.DB, period string) ([]string, []float64) {
	var sqlRead string
	var resultVals []float64
	var resultStr []string
	switch period {
	case "daily":
		sqlRead = "SELECT strftime('%d', timestamp) as valDay, SUM(amount) AS sum FROM transactions WHERE timestamp >= date('now', 'weekday 1', '-7 days') GROUP BY valDay"
	case "type":
		sqlRead = "SELECT mapping, SUM(amount) FROM transactions JOIN mappings ON mappings.description = transactions.description GROUP BY mappings.mapping"
	}

	rows, _ := db.Query(sqlRead)
	for rows.Next() {
		var day string
		var item float64
		_ = rows.Scan(&day, &item)
		resultVals = append(resultVals, item)
		resultStr = append(resultStr, day)
	}
	return resultStr, resultVals
}

func getCategories(db *sql.DB) []Category {
	var result []Category
	sqlRead := "SELECT mappings.id, mapping, description FROM transactions LEFT JOIN mappings USING (description) GROUP BY description"
	rows, _ := db.Query(sqlRead)
	for rows.Next() {
		var item Category
		_ = rows.Scan(&item.ID, &item.Mapping, &item.Description)
		result = append(result, item)
	}
	return result
}

func currentMagic(db *sql.DB) float64 {
	var magicNumber float64
	sqlRead := `SELECT
	(SELECT SUM(influence) FROM fixed) +
	(SELECT SUM(amount) FROM transactions
	WHERE datetime(timestamp) >= DATE('now'))
	AS magicnumber`
	rows := db.QueryRow(sqlRead)
	_ = rows.Scan(&magicNumber)
	return magicNumber
}
