package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/tsatke/jt"
	classpath2 "github.com/tsatke/jt/classpath"
	"github.com/tsatke/jt/jar"
)

func runFind(cmd *cobra.Command, args []string) {
	searchClass := args[0]

	project := loadProject(cwd())

	// search project and classpath concurrently, since building the classpath takes time
	wg := &sync.WaitGroup{}
	wg.Add(3)
	// search classpath
	classpathResults := make(chan string, 5)
	go func(result chan<- string) {
		defer wg.Done()
		defer close(result)

		if flagFindNoClasspath {
			// skip even building the classpath if nocp is enabled
			return
		}

		cp, err := project.Classpath()
		if err != nil {
			log.Fatal().
				Err(err).
				Str("project", project.Name()).
				Msg("get classpath")
		}

		for _, entry := range cp.Entries {
			if entry.Type != classpath2.EntryTypeJar {
				continue
			}
			jar, err := jar.Open(entry.Path)
			if err != nil {
				_, _ = fmt.Fprintln(os.Stderr, err)
				continue
			}

			for _, path := range jar.ListClasses() {
				if jt.ClassNameMatches(path, searchClass) {
					result <- path
				}
			}

			_ = jar.Close()
		}
	}(classpathResults)
	// search project files
	projectResults := make(chan string, 5)
	go func(result chan<- string) {
		defer wg.Done()
		defer close(result)

		if err := fs.WalkDir(os.DirFS("."), ".", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if d.IsDir() {
				return nil
			}

			if filepath.Ext(path) != ".java" {
				return nil
			}

			if jt.ClassNameMatches(path, searchClass) {
				result <- path
			}
			return nil
		}); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
		}
	}(projectResults)
	// print results
	go func(prjRes, cpRes <-chan string) {
		defer wg.Done()

		printWithHeader := func(header string, data <-chan string) {
			fmt.Println(header)
			for {
				res := <-data
				if res == "" {
					break
				}
				fmt.Println(res)
			}
		}

		printWithHeader("Project results:", prjRes)
		if !flagFindNoClasspath {
			printWithHeader("Classpath results:", cpRes)
		}
	}(projectResults, classpathResults)
	wg.Wait()
}
