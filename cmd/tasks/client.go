package tasks

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/Popcore/hangmango/pkg/client"
)

func init() {
	rootCmd.AddCommand(clientCmd())
}

func clientCmd() *cobra.Command {
	var port string

	cmd := &cobra.Command{
		Use:   "client",
		Short: "starts a new client session",
		Run: func(cmd *cobra.Command, args []string) {
			c, err := client.New(port)
			if err != nil {
				log.Fatal(err)
			}

			c.Play()
		},
	}
	cmd.Flags().StringVarP(&port, "port", "p", "9090", "the server port")

	return cmd
}
