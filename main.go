package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/yaml.v2"
)

var db *sql.DB

func main() {
	if len(os.Args) < 3 {
		fmt.Println("usage: " + os.Args[0] + " <workDir> <database>")
		return
	}
	workDir := os.Args[1]
	dbFile := os.Args[2]

	configBytes, err := ioutil.ReadFile(workDir + "/sqfmt.yaml")
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(configBytes, &cfg)
	if err != nil {
		panic(err)
	}

	db, err = sql.Open("sqlite3", "file:"+dbFile)
	if err != nil {
		panic(err)
	}
	raw := Data{
		Tables: queryTables(),
	}
	for group, data := range groupTables(raw) {
		saveMarkDown(workDir+"/Tabellen"+group+".md", data)
	}
}

func saveMarkDown(file string, data Data) {
	f, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	err = markdown.Execute(f, data)
	if err != nil {
		panic(err)
	}
}

func queryTables() []Table {
	tables := make([]Table, 0, 16)

	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table';")
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		name := ""
		if err := rows.Scan(&name); err != nil {
			panic(err)
		}
		tables = append(tables, queryTable(name))
	}

	return tables
}

func queryTable(name string) Table {
	table := Table{
		Name:      name,
		Cols:      make([]Col, 0, 16),
		Reference: strings.ToLower(name),
	}

	rows, err := db.Query("PRAGMA table_info(" + name + ");")
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		id := 0
		col := Col{}
		if err := rows.Scan(&id, &col.Name, &col.Type, &col.Nullable, &col.Default, &col.PrimaryKeyIndex); err != nil {
			panic(err)
		}
		table.Cols = append(table.Cols, col)
	}
	table.Name, table.Group = tableNameAndGroup(name)
	queryReferences(&table)
	return table
}

func queryReferences(table *Table) {
	rows, err := db.Query("PRAGMA foreign_key_list(" + table.Name + ");")
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		var refTableName, from string
		var id, seq, to, onUpdate, onDelete, match interface{}
		if err := rows.Scan(&id, &seq, &refTableName, &from, &to, &onUpdate, &onDelete, &match); err != nil {
			panic(err)
		}
		for i := range table.Cols {
			col := &table.Cols[i]
			if col.Name != from {
				continue
			}
			col.ReferenceTable, col.ReferenceGroup = tableNameAndGroup(refTableName)
		}
	}
}
