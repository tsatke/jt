package jt

type Project interface {
	Name() string
	Classpath() (*Classpath, error)
}

func LoadProject(path string) (Project, error) {
	switch {
	case isMavenProject(path):
		return loadMavenProject(path)
	}
	return nil, ErrUnknownProjectKind
}
