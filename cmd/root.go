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
var insecureFlag *bool
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

func failOnError(err error) {
	if err != nil {
		logrus.Fatal(err)
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logrus.Fatal(err)
	}
}

func init() {
	cobra.OnInitialize(initConfig, initLogger)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config/vecosy.yml)")
	insecureFlag = rootCmd.PersistentFlags().Bool("insecure", false, "disable the authentication")

	rootCmd.PersistentFlags().String("rest-address", ":8443", "rest address i.e. 0.0.0.0:8443")
	rootCmd.PersistentFlags().String("grpc-address", ":8080", "grpc address i.e. 0.0.0.0:8080")
	err := viper.BindPFlag("server.rest.address", rootCmd.PersistentFlags().Lookup("rest-address"))
	failOnError(err)
	err = viper.BindPFlag("server.grpc.address", rootCmd.PersistentFlags().Lookup("grpc-address"))
	failOnError(err)

	rootCmd.PersistentFlags().Bool("tls", true, "tls enabled")
	rootCmd.PersistentFlags().String("tls-cert", "", "tls certificate file")
	rootCmd.PersistentFlags().String("tls-key", "", "tls key file")
	err = viper.BindPFlag("server.tls.enabled", rootCmd.PersistentFlags().Lookup("tls"))
	failOnError(err)
	err = viper.BindPFlag("server.tls.certificateFile", rootCmd.PersistentFlags().Lookup("tls-cert"))
	failOnError(err)
	err = viper.BindPFlag("server.tls.keyFile", rootCmd.PersistentFlags().Lookup("tls-key"))
	failOnError(err)

	verboseFlag = rootCmd.Flags().BoolP("verbose", "v", false, "debug messages")
}

func initLogger() {
	if verboseFlag != nil && *verboseFlag {
		logrus.SetReportCaller(true)
		logrus.SetLevel(logrus.DebugLevel)
		logrus.SetFormatter(&logrus.TextFormatter{ForceColors: true})
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
	viper.SetDefault("security.enabled", !*insecureFlag)
}
