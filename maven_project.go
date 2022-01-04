package jt

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/spf13/afero"
	"github.com/vifraa/gopom"
)

func isMavenProject(path string) bool {
	fs := afero.NewBasePathFs(afero.NewOsFs(), path)
	return isMavenProjectFs(fs)
}

func isMavenProjectFs(fs afero.Fs) bool {
	exists, err := afero.Exists(fs, "pom.xml")
	return err == nil && exists
}

var _ Project = (*mavenProject)(nil)

type mavenProject struct {
	name string
	path string

	pom       *gopom.Project
	classpath *Classpath // nil until computed
}

func loadMavenProject(path string) (*mavenProject, error) {
	start := time.Now()

	pom, err := gopom.Parse(filepath.Join(path, "pom.xml"))
	if err != nil {
		return nil, fmt.Errorf("parse pom: %w", err)
	}

	log.Debug().
		Stringer("took", time.Since(start)).
		Msg("parse pom")

	return &mavenProject{
		name: pom.Name,
		path: path,
		pom:  pom,
	}, nil
}

func (p *mavenProject) Name() string {
	return p.name
}

func (p *mavenProject) Classpath() (*Classpath, error) {
	if p.classpath == nil {
		cp, err := p.buildClasspath()
		if err != nil {
			return nil, fmt.Errorf("build classpath: %w", err)
		}
		p.classpath = cp
	}
	return p.classpath, nil
}

func (p *mavenProject) buildClasspath() (*Classpath, error) {
	start := time.Now()

	file, err := os.CreateTemp("", "output.*")
	if err != nil {
		return nil, fmt.Errorf("create temp file: %w", err)
	}
	defer func() { _ = file.Close() }()

	buildCp := exec.Command("mvn", "dependency:build-classpath",
		"-B",
		"-q",
		"-f", filepath.Join(p.path, "pom.xml"),
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

	cp, err := ParseClasspath(buf.String())
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

			entry := &Entry{
				Type: EntryTypeJar,
				Path: filepath.Join(javaHome, path),
			}
			cp.Entries = append([]*Entry{entry}, cp.Entries...)
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
	sourceFolder := &Entry{Type: EntryTypeSource, Path: absoluteSourceDirectoryPath}
	cp.Entries = append([]*Entry{sourceFolder}, cp.Entries...)

	return cp, nil
}
