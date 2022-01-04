package main

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func runClasspath(cmd *cobra.Command, args []string) {
	project := loadProject(cwd())
	cp, err := project.Classpath()
	if err != nil {
		log.Fatal().
			Err(err).
			Str("project", project.Name()).
			Msg("get classpath")
	}
	for _, entry := range cp.Entries {
		fmt.Println(entry.Path)
	}
}
