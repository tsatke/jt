package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tsatke/jt"
)

func runClasses(cmd *cobra.Command, args []string) {
	jarFile := args[0]
	archive, err := jt.OpenJarFile(jarFile)
	if err != nil {
		log.Fatal().
			Err(err).
			Str("jar", jarFile).
			Msg("open jar file")
	}
	for _, f := range archive.ListClasses() {
		fmt.Println(f)
	}
}
