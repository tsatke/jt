package eclipse

import (
	"encoding/xml"
	"fmt"
	"io"
)

type ProjectDescription struct {
	Name      string    `xml:"name"`
	Comment   string    `xml:"comment"`
	BuildSpec BuildSpec `xml:"buildSpec"`
	Natures   Natures   `xml:"natures"`
}

type Natures struct {
	Nature []string `xml:"nature"`
}

type BuildCommand struct {
	Name string `xml:"name"`
}

type BuildSpec struct {
	BuildCommand []BuildCommand `xml:"buildCommand"`
}

func parseProjectDescription(rd io.Reader) (*ProjectDescription, error) {
	desc := &ProjectDescription{}
	err := xml.NewDecoder(rd).Decode(desc)
	if err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	return desc, nil
}
