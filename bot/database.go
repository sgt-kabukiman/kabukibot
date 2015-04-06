package bot

import "database/sql"
import _ "github.com/go-sql-driver/mysql"

type DatabaseStruct struct {
	db *sql.DB
}

func NewDatabase() *DatabaseStruct {
	return &DatabaseStruct{}
}

func (self *DatabaseStruct) Connect(dsn string) (error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}

	self.db = db

	return nil
}

func (self *DatabaseStruct) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return self.db.Query(query, args...)
}

func (self *DatabaseStruct) Exec(query string, args ...interface{}) (sql.Result, error) {
	return self.db.Exec(query, args...)
}
