// This file has been generated by `make test`.
// Do not edit it. If you need to, however, you can run `make quicktest` to
// just re-run the tests without re-creating this file.

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

func runScript(t *testing.T, filename string) {
	file, err := os.Open(filename)
	if err != nil {
		t.Error(err)
	}
	defer file.Close()

	tester := test.NewTester(file, config, db)
	initTester(tester)
	tester.WipeDatabase()
	tester.Run(t)
}

func TestAclAllow(t *testing.T) {
	runScript(t, "plugin/acl/allow.test")
}

func TestAclDeny(t *testing.T) {
	runScript(t, "plugin/acl/deny.test")
}

func TestBlacklistBasicCommands(t *testing.T) {
	runScript(t, "plugin/blacklist/basic-commands.test")
}

func TestBlacklistBasicFunctionality(t *testing.T) {
	runScript(t, "plugin/blacklist/basic-functionality.test")
}

func TestCustomCommandsAcl(t *testing.T) {
	runScript(t, "plugin/custom_commands/acl.test")
}

func TestCustomCommandsCreate(t *testing.T) {
	runScript(t, "plugin/custom_commands/create.test")
}

func TestCustomCommandsDelete(t *testing.T) {
	runScript(t, "plugin/custom_commands/delete.test")
}

func TestCustomCommandsGet(t *testing.T) {
	runScript(t, "plugin/custom_commands/get.test")
}

func TestCustomCommandsList(t *testing.T) {
	runScript(t, "plugin/custom_commands/list.test")
}

func TestCustomCommandsUpdate(t *testing.T) {
	runScript(t, "plugin/custom_commands/update.test")
}

func TestJoinJoin(t *testing.T) {
	runScript(t, "plugin/join/join.test")
}

func TestJoinLeave(t *testing.T) {
	runScript(t, "plugin/join/leave.test")
}

func TestTrollCommands(t *testing.T) {
	runScript(t, "plugin/troll/commands.test")
}

