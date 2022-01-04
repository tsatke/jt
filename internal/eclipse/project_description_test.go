package eclipse

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
)

var (
	testdataProjectDescriptions = filepath.Join("testdata", "projects")
)

func TestProjectDescriptionFileSuite(t *testing.T) {
	suite.Run(t, new(ProjectDescriptionFileSuite))
}

type ProjectDescriptionFileSuite struct {
	suite.Suite
}

func (suite *ProjectDescriptionFileSuite) TestParseProjectDescription() {
	fsys := os.DirFS(testdataProjectDescriptions)

	f, err := fsys.Open("project1")
	suite.NoError(err)
	defer func() { _ = f.Close() }()

	desc, err := parseProjectDescription(f)
	suite.NoError(err)

	suite.Equal(&ProjectDescription{
		Name:    "testproject",
		Comment: "",
		BuildSpec: BuildSpec{
			[]BuildCommand{
				{Name: "org.eclipse.jdt.core.javabuilder"},
			},
		},
		Natures: Natures{
			[]string{
				"org.eclipse.jdt.core.javanature",
			},
		},
	}, desc)
}
