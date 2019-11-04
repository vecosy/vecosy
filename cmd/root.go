package cmd

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"syscall"
)

var cfgFile string
var verboseFlag *bool

var rootCmd = &cobra.Command{
	Use:   "vecosy",
	Short: "VeCoSy - Versioned Configuration System Server",
	Run: func(cmd *cobra.Command, args []string) {
		cfgRepo := initRepo()
		go startRest(cfgRepo)
		go startGRPC(cfgRepo)
		<-waitForever()
	},
}

func waitForever() chan bool {
	done := make(chan bool, 1)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		done <- true
	}()
	logrus.Infof("stopping server")

	return done
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
		logrus.SetReportCaller(true)
		logrus.SetLevel(logrus.DebugLevel)
		logrus.SetFormatter(&logrus.TextFormatter{ForceColors: true,})
	}
}
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath("./config")
		viper.SetConfigName("vecosy")
		viper.SetConfigType("yaml")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
