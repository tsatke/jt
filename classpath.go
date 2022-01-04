package jt

import (
	"fmt"
	"strings"
	"time"
)

type Classpath struct {
	Entries []*Entry

	// classesWithLocation acts as a cache and holds fully qualified class names
	// such as java/lang/Object together with the entry in which they have been found.
	// This facilitates finding classes that were already seen in an earlier pass.
	classesWithLocation map[string]*Entry
	// cachedEntries is a slice of all entries that have been loaded into the cache field,
	// namely classesWithLocation.
	cachedEntries []*Entry
}

type Entry struct {
	Type EntryType
	Path string
}

type EntryType uint8

const (
	EntryTypeUnknown EntryType = iota
	// EntryTypeJar is used for entries that reference a jar archive.
	// Such an archive usually contains .class files.
	EntryTypeJar
	// EntryTypeSource is used for entries that reference a source folder.
	// Such a source folder usually contains .java source files.
	EntryTypeSource
)

func NewClasspath() *Classpath {
	return &Classpath{
		Entries:             nil,
		classesWithLocation: make(map[string]*Entry),
		cachedEntries:       nil,
	}
}

func ParseClasspath(cp string) (*Classpath, error) {
	start := time.Now()

	entries := strings.Split(cp, ":")
	result := NewClasspath()
	for _, entry := range entries {
		typ := EntryTypeJar
		if !strings.HasSuffix(entry, ".jar") {
			typ = EntryTypeSource
		}
		result.Entries = append(result.Entries, &Entry{
			Type: typ,
			Path: entry,
		})
	}

	log.Debug().
		Stringer("took", time.Since(start)).
		Msg("parse classpath")

	return result, nil
}

func (cp *Classpath) OpenClass(name string) (*Class, error) {
	entry := cp.classesWithLocation[name]
	if entry == nil {
		// no cache hit, find an entry that contains this class
		for _, e := range cp.Entries {
			if e.Type != EntryTypeJar {
				continue // FIXME: search in the source directory
			}
			if err := cp.loadEntryIntoCache(e); err != nil {
				return nil, fmt.Errorf("load entry: %w", err)
			}
		}
	}

	entry = cp.classesWithLocation[name]
	if entry == nil {
		// if we still get no match, that means that the class does not exist in this classpath
		return nil, nil
	}

	jar, err := OpenJarFile(entry.Path)
	if err != nil {
		return nil, fmt.Errorf("open jar file: %w", err)
	}
	defer func() { _ = jar.Close() }()

	class, err := jar.OpenClass(name)
	if err != nil {
		return nil, fmt.Errorf("open class: %w", err)
	}
	return class, nil
}

func (cp *Classpath) loadEntryIntoCache(entry *Entry) error {
	jar, err := OpenJarFile(entry.Path)
	if err != nil {
		return fmt.Errorf("open jar file: %w", err)
	}
	defer func() { _ = jar.Close() }()

	for _, className := range jar.ListClasses() {
		cp.classesWithLocation[className] = entry
	}
	cp.cachedEntries = append(cp.cachedEntries, entry)
	return nil
}
