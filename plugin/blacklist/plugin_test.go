package blacklist

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sgt-kabukiman/kabukibot/bot"
	"github.com/sgt-kabukiman/kabukibot/plugin"
	. "github.com/smartystreets/goconvey/convey"
)

func TestBlacklistPlugin(t *testing.T) {
	files, _ := filepath.Glob("*.test")

	for _, file := range files {
		Convey("Executing test script "+file, t, func() {
			file, err := os.Open(file)
			So(err, ShouldBeNil)
			defer file.Close()

			tester := plugin.NewTester(file)
			tester.AddPlugin("blacklist", func() bot.Plugin {
				return NewPlugin()
			})

			tester.Run()
		})
	}
}
