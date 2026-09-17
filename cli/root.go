package cli

import (
	cmd "fizz-buzz-rest/cli/cmd"

	cobra "github.com/spf13/cobra"
	viper "github.com/spf13/viper"
)

func Execute() error {
	root := &cobra.Command{
		Use:   "fizzbuzz",
		Short: "Fizz-buzz REST server",
		// main reports the error; cobra would otherwise print it a second time.
		SilenceErrors: true,
	}

	root.PersistentFlags().BoolP("verbose", "v", false, "enable verbose mode")
	if err := viper.BindPFlag("verbose", root.PersistentFlags().Lookup("verbose")); err != nil {
		return err
	}

	root.AddCommand(cmd.Serve())

	return root.Execute()
}
