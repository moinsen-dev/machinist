package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "machinist",
	Short: "Mac Developer Environment Snapshot & Restore CLI",
	Long:  "machinist scans your Mac developer environment and generates a restore bundle.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
