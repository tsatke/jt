package maven

import (
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"
)

func TestMavenProjectSuite(t *testing.T) {
	suite.Run(t, new(MavenProjectSuite))
}

type MavenProjectSuite struct {
	suite.Suite
}

func (suite *MavenProjectSuite) TestIsMavenProject() {
	fs := afero.NewMemMapFs()
	f, err := fs.Create("pom.xml")
	suite.NoError(err)
	suite.NoError(f.Close())

	iofs := afero.NewIOFS(fs)

	suite.True(IsMavenProjectFs(iofs))
	suite.False(IsMavenProjectFs(afero.NewIOFS(afero.NewMemMapFs())))
}

func (suite *MavenProjectSuite) TestIsMavenProjectTest1() {
	suite.True(IsMavenProject(filepath.Join("testdata", "projects", "maven", "test1")))
}

func (suite *MavenProjectSuite) TestLoadMavenProject() {
	path := filepath.Join("testdata", "projects", "maven", "test1")
	project, err := LoadProject(path)
	suite.NoError(err)
	suite.Equal("test1", project.Name())
}

func (suite *MavenProjectSuite) TestClasspath() {
	path := filepath.Join("testdata", "projects", "maven", "test1")
	project, err := LoadProject(path)
	suite.NoError(err)

	cp, err := project.Classpath()
	suite.NoError(err)

	var entries []string
	for _, entry := range cp.Entries {
		entries = append(entries, filepath.Base(entry.Path))
	}
	for _, elem := range []string{"java", "hamcrest-core-1.3.jar", "junit-4.11.jar"} {
		suite.Contains(entries, elem)
	}
}
