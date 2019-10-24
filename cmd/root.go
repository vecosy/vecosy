package cmd

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/vconf/v2/internal/rest"
	"os"
	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string
var verboseFlag *bool

var rootCmd = &cobra.Command{
	Use:   "vconf",
	Short: "VConf",
	Run: func(cmd *cobra.Command, args []string) {
		rest.StartServer()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logrus.Fatal(err)
	}
}

func init() {
	cobra.OnInitialize(initConfig, initLogger)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cmd.yaml)")
	verboseFlag = rootCmd.Flags().BoolP("verbose", "v", false, "debug messages")
}

func initLogger() {
	if verboseFlag != nil && *verboseFlag {
		logrus.SetLevel(logrus.DebugLevel)
	}
}
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		viper.AddConfigPath(home)
		viper.SetConfigName(".cmd")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
