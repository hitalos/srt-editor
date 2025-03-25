package application

import "strings"

type (
	Shortcuts []struct{ key, action string }
)

var (
	shortcuts = Shortcuts{
		{"+", "Increase time"},
		{"-", "Decrease time"},
		{"/", "Search"},
		{"n", "Search Next"},
		{"u", "Convert to UTF-8"},
	}
)

func (s Shortcuts) String() string {
	list := []string{}
	for _, sc := range s {
		list = append(list, "[yellow]"+sc.key+"[white]: "+sc.action)
	}
	return strings.Join(list, "  ")
}
