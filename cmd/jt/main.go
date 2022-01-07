package main

import (
	"context"
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/tsatke/jt"
)

var (
	root = &cobra.Command{
		Use: "jt",
	}

	subclass = &cobra.Command{
		Use:     "subclass",
		Aliases: []string{"sub"},
		Example: `Find all subclasses (slow)
jt subclass com/package/MyClass

Find all subclasses with regex (faster)
jt subclass com/package/MyClass com/package
jt subclass com/package/MyClass ^com/package
jt subclass com/package/MyClass 'com/package/.*'`,
		Short: "Print subclasses of a given class that match an optional filter",
		Long: `Print a list of all subclasses of the given class. The considered classpath is the one of the project
in the current directory. The subclasses are printed with the fully qualified name and the location.`,
		Run:  runSubclass,
		Args: cobra.RangeArgs(1, 2),
	}

	superclass = &cobra.Command{
		Use:     "superclass",
		Aliases: []string{"super"},
		Short:   "Print the parents of a given class, ending at java.lang.Object",
		Run:     runSuperclass,
		Args:    cobra.ExactArgs(1),
	}

	find = &cobra.Command{
		Use:     "find",
		Aliases: []string{"fd"},
		Short:   "Find the location of classes that match",
		Long: `Prints a list of all classes matching the argument. The list contains possible locations of the class.

If run in a terminal, this command will print headers to differentiate between matches in the project
and matches on the classpath.
If not run in a terminal (for example if the output is piped), then NO headers will be printed.`,
		Run:  runFind,
		Args: cobra.ExactArgs(1),
	}

	classpath = &cobra.Command{
		Use:     "classpath",
		Aliases: []string{"cp"},
		Short:   "Prints the classpath of the current project",
		Run:     runClasspath,
		Args:    cobra.NoArgs,
	}

	classes = &cobra.Command{
		Use:   "classes",
		Short: "Prints a list of all classes contained in the given jar file",
		Run:   runClasses,
		Args:  cobra.ExactArgs(1),
	}
)

// command line flags
var (
	verbose bool
	trace   bool

	flagFindNoClasspath bool
	flagSubclassInvert  bool
)

func init() {
	root.AddCommand(superclass, subclass, find, classpath, classes)

	root.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "print debug output")
	root.PersistentFlags().BoolVar(&trace, "trace", false, "print more debug output")
	_ = root.PersistentFlags().MarkHidden("trace")
	find.PersistentFlags().BoolVar(&flagFindNoClasspath, "no-classpath", false, "disable searching on the whole classpath and only search in the project")
	subclass.PersistentFlags().BoolVar(&flagSubclassInvert, "invert", false, "invert the matching, considering all classes that don't match the pattern")
}

func main() {
	for _, command := range root.Commands() {
		_ = command.ParseFlags(os.Args)
	}

	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).
		With().
		Timestamp().
		Logger().
		Level(zerolog.ErrorLevel)

	if verbose {
		log.Logger = log.Logger.Level(zerolog.DebugLevel)
	}
	if trace {
		log.Logger = log.Logger.Level(zerolog.TraceLevel)
	}

	ctx := context.Background()
	if err := root.ExecuteContext(ctx); err != nil {
		panic(err)
	}
}

func loadProject(path string) jt.Project {
	project, err := jt.LoadProject(path)
	if err != nil {
		log.Fatal().
			Err(err).
			Str("path", path).
			Msg("load project")
	}

	return project
}

func cwd() string {
	cwd, err := filepath.Abs(".")
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("get cwd")
	}
	return cwd
}
