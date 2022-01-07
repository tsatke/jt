package main

import (
	"fmt"
	"regexp"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func runSubclass(cmd *cobra.Command, args []string) {
	classname := args[0]
	regex := ""
	if len(args) > 1 {
		regex = args[1]
	}

	project := loadProject(cwd())
	classpath, err := project.Classpath()
	if err != nil {
		log.Fatal().
			Err(err).
			Str("project", project.Name()).
			Msg("get classpath")
	}

	pattern := regexp.MustCompile(regex)

	resultsCh := make(chan string, 5)
	go classpath.FindClasses(func(s string) bool {
		condition := regex != "" && !pattern.MatchString(s)
		if flagSubclassInvert {
			condition = !condition
		}
		if condition {
			return false
		}

		c, err := classpath.OpenClass(s)
		if err != nil {
			return false
		}
		return c.SuperclassName() == classname
	}, resultsCh)

	for ch := range resultsCh {
		fmt.Println(ch)
	}
}
