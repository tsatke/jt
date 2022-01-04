package main

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func runSuperclass(cmd *cobra.Command, args []string) {
	classname := args[0]

	project := loadProject(cwd())
	classpath, err := project.Classpath()
	if err != nil {
		log.Fatal().
			Err(err).
			Str("project", project.Name()).
			Msg("get classpath")
	}

	for classname != "" {
		fmt.Println(classname)
		class, err := classpath.OpenClass(classname)
		if err != nil {
			log.Fatal().
				Err(err).
				Str("project", project.Name()).
				Str("class", classname).
				Msg("open class")
		}
		if class == nil {
			log.Fatal().
				Err(err).
				Str("project", project.Name()).
				Str("class", classname).
				Msg("class not on classpath")
		}
		classname = class.SuperclassName()
	}
}
