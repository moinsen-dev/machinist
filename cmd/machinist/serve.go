package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var servePort int

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run as MCP server",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintln(cmd.OutOrStdout(), "serve: not yet implemented")
	},
}

func init() {
	serveCmd.Flags().IntVarP(&servePort, "port", "p", 8080, "Port to listen on")
	rootCmd.AddCommand(serveCmd)
}
