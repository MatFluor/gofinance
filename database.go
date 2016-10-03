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

func currentMagic(db *sql.DB) float64 {
	var magicNumber float64
	sqlRead := `SELECT
	(SELECT SUM(influence) FROM fixed) +
	(SELECT SUM(amount) FROM transactions
	WHERE datetime(timestamp) >= DATE('now'))
	AS magicnumber`
	rows, err := db.Query(sqlRead)
	defer rows.Close()
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		_ = rows.Scan(&magicNumber)
	}
	return magicNumber
}
