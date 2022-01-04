package eclipse

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
)

var (
	testdataClasspaths = filepath.Join("testdata", "classpaths")
)

func TestClasspathFileSuite(t *testing.T) {
	suite.Run(t, new(ClasspathFileSuite))
}

type ClasspathFileSuite struct {
	suite.Suite
}

func (suite *ClasspathFileSuite) TestParseClasspath() {
	fsys := os.DirFS(testdataClasspaths)

	f, err := fsys.Open("classpath1")
	suite.NoError(err)
	defer func() { _ = f.Close() }()

	cp, err := parseClasspathFile(f)
	suite.NoError(err)

	entries := cp.Entries
	// order matters here!
	suite.Equal([]ClasspathEntry{
		{Kind: "src", Path: "src/java", Including: "**/*.java"},
		{Kind: "src", Path: "src/res", Including: "**/*", Excluding: "**/*.java"},
		{Kind: "src", Path: "test/java", Output: "target/test-classes", Including: "**/*.java"},
		{Kind: "src", Path: "test/res", Output: "target/test-classes", Excluding: "**/.svn/**|**/*.java"},
		{Kind: "output", Path: "target/classes"},
		{Kind: "con", Path: "org.eclipse.jdt.launching.JRE_CONTAINER"},
		{Kind: "var", Path: "M2_REPO/junit/junit/4.12/junit-4.12.jar", Sourcepath: "M2_REPO/junit/junit/4.12/junit-4.12-sources.jar"},
		{Kind: "var", Path: "M2_REPO/org/hamcrest/hamcrest-core/2.2/hamcrest-core-2.2.jar", Sourcepath: "M2_REPO/org/hamcrest/hamcrest-core/2.2/hamcrest-core-2.2-sources.jar"},
	}, entries)
}
