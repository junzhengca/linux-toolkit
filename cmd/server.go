package cmd

import (
	"fmt"
	"os"

	"github.com/jun/linux-toolkit/pkg/server"
	"github.com/spf13/cobra"
)

var (
	serverPort     int
	serverBind     string
	serverInterval int
	serverNoUI     bool
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start HTTP server for web access",
	Long:  `Start HTTP server that exposes system metrics as JSON API with HTML UI`,
	Run:   runServer,
}

func runServer(cmd *cobra.Command, args []string) {
	srv := server.NewServer(serverPort, serverBind, !serverNoUI, serverInterval)
	if err := srv.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to start server: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	serverCmd.Flags().IntVarP(&serverPort, "port", "p", 8080, "Port number")
	serverCmd.Flags().StringVarP(&serverBind, "bind", "b", "0.0.0.0", "Bind address")
	serverCmd.Flags().IntVarP(&serverInterval, "interval", "i", 5, "UI refresh interval in seconds")
	serverCmd.Flags().BoolVarP(&serverNoUI, "no-ui", "n", false, "Start API only, no HTML UI")
}
