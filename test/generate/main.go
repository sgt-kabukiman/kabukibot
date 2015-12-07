package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

type testcase struct {
	Filename string
	Function string
}

func main() {
	root := "plugin"
	exploder := regexp.MustCompile(`[^a-z0-9]`)
	testcases := make([]testcase, 0)

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || !strings.HasSuffix(path, ".test") {
			return nil
		}

		rel, _ := filepath.Rel(root, path)
		rel = filepath.ToSlash(rel)

		name := strings.TrimSuffix(rel, ".test")
		parts := exploder.Split(name, -1)

		for i := range parts {
			parts[i] = strings.ToUpper(string(parts[i][0])) + parts[i][1:]
		}

		testcases = append(testcases, testcase{
			Filename: root + "/" + rel,
			Function: strings.Join(parts, ""),
		})

		return nil
	})

	filename := root + "/../plugin_test.go"
	fp, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0660)
	if err != nil {
		panic(err)
	}

	cTemplate, _ := template.ParseFiles("test/generate/testfile.got")
	cTemplate.Execute(fp, map[string]interface{}{
		"testcases": testcases,
	})

	fp.Close()
}
