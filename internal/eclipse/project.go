package eclipse

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/tsatke/jt/classpath"
)

const (
	ClasspathFileName = ".classpath"
	ProjectFileName   = ".project"
)

func IsEclipseProject(path string) bool {
	return IsEclipseProjectFs(os.DirFS(path))
}

func IsEclipseProjectFs(fsys fs.FS) bool {
	cpInfo, err := fs.Stat(fsys, ClasspathFileName)
	if err != nil || cpInfo == nil || cpInfo.IsDir() {
		return false
	}

	projectInfo, err := fs.Stat(fsys, ProjectFileName)
	if err != nil || projectInfo == nil || projectInfo.IsDir() {
		return false
	}

	return true
}

type project struct {
	path string

	classpathFile      *ClasspathFile
	projectDescription *ProjectDescription

	classpath *classpath.Classpath // nil until computed
}

func LoadProject(path string) (*project, error) {
	start := time.Now()

	fsys := os.DirFS(path)
	projectFile, err := fsys.Open(ProjectFileName)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", ProjectFileName, err)
	}
	defer func() { _ = projectFile.Close() }()

	projectDescription, err := parseProjectDescription(projectFile)
	if err != nil {
		return nil, fmt.Errorf("parse project description: %w", err)
	}

	classpathFile, err := fsys.Open(ClasspathFileName)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", ClasspathFileName, err)
	}
	defer func() { _ = classpathFile.Close() }()

	cp, err := parseClasspathFile(classpathFile)
	if err != nil {
		return nil, fmt.Errorf("parse classpath file: %w", err)
	}

	log.Debug().
		Stringer("took", time.Since(start)).
		Msg("parse classpath and project description")

	return &project{
		path:               path,
		classpathFile:      cp,
		projectDescription: projectDescription,
	}, nil
}

func (p *project) Name() string {
	return p.projectDescription.Name
}

func (p *project) Classpath() (*classpath.Classpath, error) {
	if p.classpath == nil {
		cp, err := p.buildClasspath()
		if err != nil {
			return nil, err
		}
		p.classpath = cp
	}

	return p.classpath, nil
}

func (p *project) buildClasspath() (*classpath.Classpath, error) {
	cp := classpath.NewClasspath()
	for _, entry := range p.classpathFile.Entries {
		switch entry.Kind { // TODO: handle 'src' (if necessary) and 'con'
		case "output":
			path, err := filepath.Abs(entry.Path)
			if err != nil {
				return nil, fmt.Errorf("make path absolute:%w", err)
			}
			cp.AddEntry(classpath.EntryTypeOutput, path)
		case "var":
			fragments := strings.Split(entry.Path, "/")
			if resolvedVariable := os.Getenv(fragments[0]); resolvedVariable != "" {
				fragments[0] = resolvedVariable
			}
			path := filepath.Join(fragments...)
			cp.AddEntry(classpath.EntryTypeJar, path)
		}
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

	return cp, nil
}
