package maven

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/tsatke/jt/classpath"
	"github.com/vifraa/gopom"
)

const (
	PomFileName = "pom.xml"
)

func IsMavenProject(path string) bool {
	return IsMavenProjectFs(os.DirFS(path))
}

func IsMavenProjectFs(fsys fs.FS) bool {
	pomInfo, err := fs.Stat(fsys, PomFileName)
	return err == nil && pomInfo != nil && !pomInfo.IsDir()
}

type project struct {
	path string

	pom       *gopom.Project
	classpath *classpath.Classpath // nil until computed
}

func LoadProject(path string) (*project, error) {
	start := time.Now()

	pom, err := gopom.Parse(filepath.Join(path, PomFileName))
	if err != nil {
		return nil, fmt.Errorf("parse pom: %w", err)
	}

	log.Debug().
		Stringer("took", time.Since(start)).
		Msg("parse pom")

	return &project{
		path: path,
		pom:  pom,
	}, nil
}

func (p *project) Name() string {
	return p.pom.Name
}

func (p *project) Classpath() (*classpath.Classpath, error) {
	if p.classpath == nil {
		cp, err := p.buildClasspath()
		if err != nil {
			return nil, fmt.Errorf("build classpath: %w", err)
		}
		p.classpath = cp
	}
	return p.classpath, nil
}

func (p *project) buildClasspath() (*classpath.Classpath, error) {
	start := time.Now()

	file, err := os.CreateTemp("", "output.*")
	if err != nil {
		return nil, fmt.Errorf("create temp file: %w", err)
	}
	defer func() { _ = file.Close() }()

	// TODO: it seems like maven has some kind of caching mechanism, and if regenerateFile is not true, it will not write output if the project hasn't changed
	buildCp := exec.Command("mvn", "dependency:build-classpath",
		"-B",
		"-q",
		"-f", filepath.Join(p.path, PomFileName),
		"-Dmdep.outputFile="+file.Name(),
		"-Dmdep.regenerateFile=true",
	)
	data, err := buildCp.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("build classpath: %w (%s)", err, string(data))
	}
	log.Trace().
		Stringer("command", buildCp).
		Msg("build classpath with command")

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(file)
	if err != nil {
		return nil, fmt.Errorf("read classpath: %w", err)
	}

	log.Debug().
		Stringer("took", time.Since(start)).
		Msg("build classpath")

	cp, err := classpath.Parse(buf.String())
	if err != nil {
		return nil, fmt.Errorf("parse classpath: %w", err)
	}

	// add JAVA_HOME at the beginning of the classpath
	javaHome := os.Getenv("JAVA_HOME")
	if javaHome != "" {
		if err := fs.WalkDir(os.DirFS(javaHome), ".", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if d.IsDir() {
				return nil
			}

			if filepath.Ext(path) != ".jar" {
				return nil
			}

			entry := &classpath.Entry{
				Type: classpath.EntryTypeJar,
				Path: filepath.Join(javaHome, path),
			}
			cp.Entries = append([]*classpath.Entry{entry}, cp.Entries...)
			return nil
		}); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
		}
	}

	// add the maven project source folder at the beginning of the classpath
	sourceDirectoryPath := p.pom.Build.SourceDirectory
	if sourceDirectoryPath == "" {
		// if no source directory path is set, use the maven default as a fallback
		sourceDirectoryPath = "src/main/java" // TODO: make this a constant
	}
	absoluteSourceDirectoryPath, err := filepath.Abs(sourceDirectoryPath)
	if err != nil {
		return nil, fmt.Errorf("unable to make source directory path absolute: %w", err)
	}
	sourceFolder := &classpath.Entry{Type: classpath.EntryTypeSource, Path: absoluteSourceDirectoryPath}
	cp.Entries = append([]*classpath.Entry{sourceFolder}, cp.Entries...)

	return cp, nil
}
