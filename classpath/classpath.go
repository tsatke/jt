package classpath

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/tsatke/jt/class"
	"github.com/tsatke/jt/jar"
)

type Classpath struct {
	Entries []*Entry

	// classesWithLocation acts as a cache and holds fully qualified class names
	// such as java/lang/Object together with the entry in which they have been found.
	// This facilitates finding classes that were already seen in an earlier pass.
	classesWithLocation map[string]*Entry
	// cachedEntries is a set of all entries that have been loaded into the cache field,
	// namely classesWithLocation.
	cachedEntries map[string]struct{}
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
	// EntryTypeOutput is used for entries that reference a folder containing
	// compiled sources. These compiled sources usually are .class files.
	EntryTypeOutput
)

func NewClasspath() *Classpath {
	return &Classpath{
		Entries:             nil,
		classesWithLocation: make(map[string]*Entry),
		cachedEntries:       make(map[string]struct{}),
	}
}

func Parse(cp string) (*Classpath, error) {
	entries := filepath.SplitList(cp)
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

	return result, nil
}

func (cp *Classpath) AddEntry(typ EntryType, path string) {
	cp.Entries = append(cp.Entries, &Entry{typ, path})
}

func (cp *Classpath) OpenClass(name string) (*class.Class, error) {
	return cp.OpenClassWithCache(name, nil)
}

func (cp *Classpath) OpenClassWithCache(name string, cache *jar.Cache) (*class.Class, error) {
	entry := cp.classesWithLocation[name]
	if entry == nil {
		// no cache hit, find an entry that contains this class
		for _, e := range cp.Entries {

			// only search entries that are not loaded into the cache yet
			if _, ok := cp.cachedEntries[e.Path]; ok {
				continue
			}

			if e.Type != EntryTypeJar {
				continue // FIXME: search in the source directory
			}
			if err := cp.loadEntryIntoCache(e); err != nil {
				return nil, fmt.Errorf("load entry: %w", err)
			}

			// we cached an entry that contains the class we are looking for
			if cp.classesWithLocation[name] != nil {
				break
			}
		}
		entry = cp.classesWithLocation[name]
	}

	if entry == nil {
		// if we still get no match, that means that the class does not exist in this classpath
		return nil, nil
	}

	log.Trace().
		Str("entry", entry.Path).
		Str("search", name).
		Msg("found match")

	var jf *jar.File
	var err error
	if cache == nil {
		jf, err = jar.Open(entry.Path)
		if err != nil {
			return nil, fmt.Errorf("open jar file: %w", err)
		}
		defer func() { _ = jf.Close() }()
	} else {
		var ok bool
		jf, ok = cache.Get(entry.Path)
		if !ok {
			jf, err = jar.Open(entry.Path)
			if err != nil {
				return nil, fmt.Errorf("open jar file: %w", err)
			}
			cache.Add(entry.Path, jf)
		}
	}

	class, err := jf.OpenClass(name)
	if err != nil {
		return nil, fmt.Errorf("open class: %w", err)
	}
	return class, nil
}

func (cp *Classpath) FindClasses(matchFn func(string) bool, resultsCh chan<- string) {
	start := time.Now()

	// first, load all entries into the cache
	for _, e := range cp.Entries {

		// only search entries that are not loaded into the cache yet
		if _, ok := cp.cachedEntries[e.Path]; ok {
			continue
		}

		if e.Type != EntryTypeJar {
			continue // FIXME: search in the source directory
		}
		if err := cp.loadEntryIntoCache(e); err != nil {
			close(resultsCh)
			return
		}
	}

	log.Debug().
		Stringer("took", time.Since(start)).
		Int("classes", len(cp.classesWithLocation)).
		Msg("load classpath")

	start = time.Now()

	// then, search for the class that matches
	sourceCh := make(chan string, 50)
	go func() {
		for classname, _ := range cp.classesWithLocation {
			sourceCh <- classname
		}
		close(sourceCh)
	}()

	routines := runtime.NumCPU()
	if routines < 1 {
		routines = 1
	}

	log.Debug().
		Int("routines", routines).
		Msg("searching concurrently")

	wg := &sync.WaitGroup{}
	for i := 0; i < routines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				in, ok := <-sourceCh
				if !ok {
					break
				}
				if matchFn(in) {
					resultsCh <- in
				}
			}
		}()
	}
	wg.Wait()

	log.Debug().
		Stringer("took", time.Since(start)).
		Int("classes", len(cp.classesWithLocation)).
		Msg("find class")

	close(resultsCh)
}

func (cp *Classpath) loadEntryIntoCache(entry *Entry) error {
	start := time.Now()

	jar, err := jar.Open(entry.Path)
	if err != nil {
		return fmt.Errorf("open jar file: %w", err)
	}
	defer func() { _ = jar.Close() }()

	classes := 0
	for _, className := range jar.ListClasses() {
		// since we work through the classpath top to bottom, don't overwrite entries
		if cp.classesWithLocation[className] == nil {
			cp.classesWithLocation[className] = entry
			classes++
		} else {
			if log.Trace().Enabled() {
				log.Warn().
					Str("class", className).
					Str("original", cp.classesWithLocation[className].Path).
					Str("duplicate", entry.Path).
					Msg("found in two entries, discarded duplicate")
			}
		}
	}
	cp.cachedEntries[entry.Path] = struct{}{}

	log.Trace().
		Stringer("took", time.Since(start)).
		Str("entry", entry.Path).
		Int("classes", classes).
		Msg("load entry into cache")

	return nil
}
