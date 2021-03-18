package funcs

import (
	"bytes"
	"fmt"
	"strings"
)

func DumpSupports(showDesc, showDemo bool) string {

	s := bytes.NewBufferString("")

	for _, f := range SupportFuncs {
		s.WriteString(fmt.Sprintf("%s", strings.TrimSpace(f.Title)))
		s.WriteString("\n")
		if showDesc {
			s.WriteString(fmt.Sprintf("  %s", strings.TrimSpace(f.Desc)))
			s.WriteString("\n\n")
		}

		if showDemo {
			for idx, t := range f.Test {
				s.WriteString(fmt.Sprintf("  Demo %d:", idx+1))
				s.WriteString(t)
			}
			s.WriteString("\n\n")
		}

	}
	fmt.Printf("%s\n", s.String())
	return s.String()
}
