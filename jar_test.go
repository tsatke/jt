package jt

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestJarSuite(t *testing.T) {
	suite.Run(t, new(JarSuite))
}

type JarSuite struct {
	suite.Suite
}

func (suite *JarSuite) TestOpenClass() {
	jar, err := OpenJarFile(filepath.Join("testdata", "jars", "test1.jar"))
	suite.NoError(err)

	class, err := jar.OpenClass("com/github/tsatke/jt/App")
	suite.NoError(err)
	suite.Equal("com/github/tsatke/jt/App", class.Name())
	suite.Equal("java/lang/Object", class.SuperclassName())

	methods := class.Methods()
	suite.Len(methods, 2)
	suite.Equal("<init>", methods[0].Name())
	suite.Equal("main", methods[1].Name())
}

func (suite *JarSuite) TestListClasses() {
	jar, err := OpenJarFile(filepath.Join("testdata", "jars", "test1.jar"))
	suite.NoError(err)

	suite.ElementsMatch([]string{"com/github/tsatke/jt/App"}, jar.ListClasses())
}
