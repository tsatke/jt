package jt

import (
	"github.com/tsatke/jt/classpath"
	"github.com/tsatke/jt/maven"
)

type Project interface {
	Name() string
	Classpath() (*classpath.Classpath, error)
}

func LoadProject(path string) (Project, error) {
	switch {
	case maven.IsMavenProject(path):
		return maven.LoadProject(path)
	}
	return nil, ErrUnknownProjectKind
}
