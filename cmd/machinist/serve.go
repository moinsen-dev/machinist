package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	mcpserver "github.com/moinsen-dev/machinist/internal/mcp"

	"github.com/mark3labs/mcp-go/server"
)

var servePort int

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run as MCP server",
	RunE: func(cmd *cobra.Command, args []string) error {
		reg := newRegistry()
		srv := mcpserver.NewMachinistServer(reg)

		port, _ := cmd.Flags().GetInt("port")
		if port > 0 {
			// SSE transport
			return server.NewSSEServer(srv.MCPServer(),
				server.WithBaseURL(fmt.Sprintf("http://localhost:%d", port)),
			).Start(fmt.Sprintf(":%d", port))
		}
		// Default: stdio transport
		return server.NewStdioServer(srv.MCPServer()).Listen(context.Background(), os.Stdin, os.Stdout)
	},
}

func init() {
	serveCmd.Flags().IntVarP(&servePort, "port", "p", 0, "Port to listen on (0 = stdio)")
	rootCmd.AddCommand(serveCmd)
}
