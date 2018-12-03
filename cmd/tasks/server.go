package tasks

import (
	"github.com/spf13/cobra"

	"github.com/Popcore/hangmango/pkg/server"
)

func init() {
	rootCmd.AddCommand(serverCmd())
}

func serverCmd() *cobra.Command {
	var port string
	var verbose bool

	cmd := &cobra.Command{
		Use:   "server",
		Short: "starts a new game server",
		Run: func(cmd *cobra.Command, args []string) {
			s := server.New(port, verbose)
			s.Start()
		},
	}
	cmd.Flags().StringVarP(&port, "port", "p", "9090", "the server port")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", true, "print logs to stout")

	return cmd
}
