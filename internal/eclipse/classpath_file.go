package eclipse

import (
	"encoding/xml"
	"fmt"
	"io"
)

type ClasspathFile struct {
	Entries []ClasspathEntry `xml:"classpathentry"`
}

type ClasspathEntry struct {
	Kind       string `xml:"kind,attr"`
	Path       string `xml:"path,attr"`
	Output     string `xml:"output,attr"`
	Including  string `xml:"including,attr"`
	Excluding  string `xml:"excluding,attr"`
	Sourcepath string `xml:"sourcepath,attr"`
}

func parseClasspathFile(rd io.Reader) (*ClasspathFile, error) {
	cp := &ClasspathFile{}
	err := xml.NewDecoder(rd).Decode(cp)
	if err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	return cp, nil
}
