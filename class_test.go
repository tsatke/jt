package jt

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestClassSuite(t *testing.T) {
	suite.Run(t, new(ClassSuite))
}

type ClassSuite struct {
	suite.Suite
}

func (suite *ClassSuite) TestParse() {
	f, err := os.Open(filepath.Join("testdata", "classes", "App.class"))
	suite.NoError(err)
	defer func() { _ = f.Close() }()

	res, err := ParseClass(f)
	suite.NoError(err)

	suite.Equal("com/github/tsatke/jt/App", res.Name())
	suite.Equal("java/lang/Object", res.SuperclassName())

	methods := res.Methods()
	suite.Len(methods, 2)
	suite.Equal("<init>", methods[0].Name())
	suite.Equal("main", methods[1].Name())
}
