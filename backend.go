package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"io/ioutil"
	"path"
	"strings"
)

const backendDir = "backend"

type taxiiConnector interface {
	connect() error
	disconnect()
}

type taxiiWriter interface {
	create(stmt string) error
	delete(stmt string) error
	update(stmt string) error
}

type taxiiReader interface {
	read(stmt string) error
}

type taxiiDataStore interface {
	taxiiConnector
	taxiiReader
	taxiiWriter
}

type sqliteDB struct {
	db       *sql.DB
	language string
	path     string
	driver   string
}

/* sqlite */

func newSQLiteDB() sqliteDB {
	config := cabbyConfig{}.parse(configPath)
	s := sqliteDB{language: "sql", driver: "sqlite3", path: config.DataStore["path"]}
	return s
}

/* connector methods */

func (s *sqliteDB) connect(path string) error {
	logInfo.Println("Connecting to", path)

	var err error
	s.db, err = sql.Open("sqlite3", path)
	return err
}

func (s *sqliteDB) disconnect() {
	s.db.Close()
}

/* writer methods */

func (s *sqliteDB) create(container string, args map[string]string) error {
	statement := parseStatement(s.language, "create", container)
	statement = swapArgs(statement, args)

	_, err := s.db.Exec(statement)
	if err != nil {
		logError.Printf("Error: %v in statement \"%v\"\n", err, statement)
	}
	return err
}

/* helpers */

func parseStatement(language, command, container string) string {
	fileName := strings.Join([]string{command, container}, "_")
	fileName = fileName + "." + language
	path := path.Join(backendDir, language, fileName)

	statement, err := ioutil.ReadFile(path)
	if err != nil {
		logError.Panicf("Unable to parse statement file: %v", path)
	}

	return string(statement)
}

func swapArgs(statement string, args map[string]string) string {
	for k, v := range args {
		statement = strings.Replace(statement, "$"+k, v, -1)
	}
	return statement
}
