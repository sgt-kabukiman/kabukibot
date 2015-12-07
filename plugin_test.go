package main

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/sgt-kabukiman/kabukibot/bot"
	"github.com/sgt-kabukiman/kabukibot/test"
)

var config *bot.Configuration
var db *sqlx.DB

func init() {
	var err error

	// load configuration
	config, err = bot.LoadConfiguration("config-test.yaml")
	if err != nil {
		panic(err)
	}

	// connect to database
	db, err = sqlx.Connect("mysql", config.Database.DSN)
	if err != nil {
		panic(err)
	}
}

func TestBlacklistBasicCommands(t *testing.T) {
	file, err := os.Open("plugin/blacklist/basic-commands.test")
	if err != nil {
		t.Error(err)
	}
	defer file.Close()

	tester := test.NewTester(file, config, db)
	initTester(tester)
	tester.WipeDatabase()
	tester.Run(t)
}

func TestBlacklistBasicFunctionality(t *testing.T) {
	file, err := os.Open("plugin/blacklist/basic-functionality.test")
	if err != nil {
		t.Error(err)
	}
	defer file.Close()

	tester := test.NewTester(file, config, db)
	initTester(tester)
	tester.WipeDatabase()
	tester.Run(t)
}

func TestJoin(t *testing.T) {
	file, err := os.Open("plugin/join.test")
	if err != nil {
		t.Error(err)
	}
	defer file.Close()

	tester := test.NewTester(file, config, db)
	initTester(tester)
	tester.WipeDatabase()
	tester.Run(t)
}
