package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var config Config

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "rcm-node-installer",
	Short: "rcm-node-installer manages containerd shims",
	PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
		return initializeConfig(cmd)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&config.Runtime.Name, "runtime", "r", "containerd", "Set the container runtime to configure (containerd, cri-o)")
	rootCmd.PersistentFlags().StringVarP(&config.Runtime.ConfigPath, "runtime-config", "c", "", "Path to the runtime config file. Will try to autodetect if left empty")
	rootCmd.PersistentFlags().StringVarP(&config.RCM.Path, "rcm-path", "k", "/opt/rcm", "Working directory for the RuntimeClassManager on the host")
	rootCmd.PersistentFlags().StringVarP(&config.Host.RootPath, "host-root", "H", "/", "Path to the host root path")
}

func initializeConfig(cmd *cobra.Command) error {
	v := viper.New()

	v.SetConfigName("rcm")

	v.AddConfigPath(".")
	v.AddConfigPath("/etc")

	if err := v.ReadInConfig(); err != nil {
		if !errors.As(err, &viper.ConfigFileNotFoundError{}) {
			return err
		}
	}

	// Environment variables are prefixed with RCM_
	v.SetEnvPrefix("RCM")

	// - is replaced with _ in environment variables
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	// Bind env variables
	v.AutomaticEnv()

	bindFlags(cmd, v)

	return nil
}

// bindFlags binds each cobra flag to its associated viper configuration.
func bindFlags(cmd *cobra.Command, v *viper.Viper) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		configName := strings.ReplaceAll(f.Name, "-", "_")

		if !f.Changed && v.IsSet(configName) {
			val := v.Get(configName)
			_ = cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
		}
	})
}
