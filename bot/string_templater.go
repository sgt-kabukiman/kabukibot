package bot

import "fmt"
import "regexp"
import "strings"
import "time"

type StringTemplater interface {
	Render(string) string
}

type stringRenderer    func(string) string
type stringRendererMap map[string]stringRenderer

type stringTemplater struct {
	renderers stringRendererMap
}

func NewStringTemplater() StringTemplater {
	templater := &stringTemplater{make(stringRendererMap)}

	templater.AddRenderer("reldate", func(dateString string) string {
		// try to parse the date
		t, err := time.Parse("2 Jan. 2006", dateString)
		if err != nil {
			return dateString
		}

		now      := time.Now()
		duration := int(now.Sub(t).Hours() / 24)

		switch duration {
			case 0:  return "today"
			case 1:  return "yesterday"
			case -1: return "tomorrow"
		}

		if duration > 0 {
			return fmt.Sprintf("%d days ago", duration)
		} else {
			return fmt.Sprintf("in %d days", -duration)
		}
	})

	return templater
}

func (self *stringTemplater) AddRenderer(section string, r stringRenderer) {
	self.renderers[section] = r
}

func (self *stringTemplater) RemoveRenderer(section string) {
	delete(self.renderers, section)
}

var tplSectionRegex = regexp.MustCompile(`<([a-z0-9]+)>([^%]*)</>`)

func (self *stringTemplater) Render(input string) string {
	return tplSectionRegex.ReplaceAllStringFunc(input, func(match string) string {
		parts := tplSectionRegex.FindStringSubmatch(match)
		if len(parts) == 0 { // should never happen
			return match
		}

		section := strings.ToLower(parts[1])
		content := strings.TrimSpace(parts[2])

		renderer, exists := self.renderers[section]
		if !exists {
			return content
		}

		return renderer(content)
	})
}
