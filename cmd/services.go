package cmd

import (
	"fmt"
	"os"

	"github.com/jun/linux-toolkit/pkg/services"
	"github.com/spf13/cobra"
)

var (
	serviceName     string
	serviceStatus   string
	serviceUser     string
	servicePort     int
	servicesFormat  string
	showPorts       bool
	showAllServices bool
	systemdOnly     bool
)

var servicesCmd = &cobra.Command{
	Use:   "services",
	Short: "Display services running on the system",
	Long: `Display detailed information about services running on the system,
including daemons and web services that bind to ports. Shows service name,
status, PID, memory usage, CPU usage, start time, uptime, user, command,
and listening ports.`,
	Run: func(cmd *cobra.Command, args []string) {
		reader := services.NewServicesReader()

		// If service name is specified, show detailed info for that service
		if serviceName != "" {
			info, err := reader.ReadServiceByName(serviceName)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: Failed to read service '%s': %v\n", serviceName, err)
				os.Exit(1)
			}
			services.PrintDetail(info)
			return
		}

		// Read all services
		summary, err := reader.ReadAllServices()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to read services: %v\n", err)
			os.Exit(1)
		}

		// Apply filters
		filteredServices := reader.FilterServices(
			summary.Services,
			serviceName,
			serviceStatus,
			serviceUser,
			servicePort,
		)
		summary.Services = filteredServices
		summary.TotalServices = len(filteredServices)

		// Recalculate summary after filtering
		summary.Running = 0
		summary.Stopped = 0
		summary.Failed = 0
		for _, s := range filteredServices {
			switch s.Status {
			case "running":
				summary.Running++
			case "stopped":
				summary.Stopped++
			case "failed":
				summary.Failed++
			}
		}

		// Parse output format
		format := services.FormatTable
		if servicesFormat == "json" {
			format = services.FormatJSON
		}

		// Set print options
		opts := services.PrintOptions{
			ShowPorts:  showPorts,
			ShowAll:    showAllServices,
			ShowDetail: false,
		}

		// Print services information
		if err := services.Print(summary, format, opts); err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to print services: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(servicesCmd)

	servicesCmd.Flags().StringVarP(&serviceName, "name", "n", "", "Filter by service name (supports partial match)")
	servicesCmd.Flags().StringVarP(&serviceStatus, "status", "s", "", "Filter by status (running|stopped|failed)")
	servicesCmd.Flags().StringVarP(&serviceUser, "user", "u", "", "Filter by user")
	servicesCmd.Flags().IntVarP(&servicePort, "port", "p", 0, "Filter by port number")
	servicesCmd.Flags().BoolVar(&showPorts, "show-ports", false, "Show listening ports for each service")
	servicesCmd.Flags().BoolVarP(&showAllServices, "show-all", "a", false, "Show all services including stopped ones")
	servicesCmd.Flags().BoolVar(&systemdOnly, "systemd-only", false, "Show only systemd services")
	servicesCmd.Flags().StringVarP(&servicesFormat, "format", "f", "table", "Output format (table|json)")
}
