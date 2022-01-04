package main

import (
	"github.com/spf13/cobra"
	"github.com/tsatke/jt"
)

func runSubclass(cmd *cobra.Command, args []string) {
	rootClass := args[0]
	_ = rootClass
	panic(jt.ErrNotImplemented)
}
