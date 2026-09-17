package cmd

import (
	"log/slog"
	"os"

	server "fizz-buzz-rest/server"
	config "fizz-buzz-rest/utils/config"

	cobra "github.com/spf13/cobra"
	viper "github.com/spf13/viper"
)

func Serve() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Launch server",
	}

	cmd.Flags().Int("port", 0, "port to listen on, overriding APP_PORT")

	cmd.Run = func(cmd *cobra.Command, args []string) {
		if cmd.Flags().Changed("port") {
			port, err := cmd.Flags().GetInt("port")
			if err != nil {
				slog.Error("reading the port flag failed", "error", err)
				os.Exit(1)
			}
			viper.Set("APP_PORT", port)
		}

		if err := config.Load(); err != nil {
			slog.Error("loading the configuration failed", "error", err)
			os.Exit(1)
		}

		if err := server.Launch(); err != nil {
			slog.Error("server stopped", "error", err)
			os.Exit(1)
		}
	}

	return cmd
}
