package cmd

import (
	"fmt"
	"github.com/vecosy/vecosy/v2/internal/rest"
	"github.com/vecosy/vecosy/v2/pkg/configrepo/configGitRepo"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/src-d/go-git.v4"
	"path"
	"path/filepath"
	"time"
)

var cfgFile string
var verboseFlag *bool

var rootCmd = &cobra.Command{
	Use:   "vconf",
	Short: "VConf",
	Run: func(cmd *cobra.Command, args []string) {
		repoUrl := viper.GetString("repo.url")
		if !path.IsAbs(repoUrl) {
			repoUrl, _ = filepath.Abs(repoUrl)
		}
		cfgRepo, err := configGitRepo.NewConfigRepo(viper.GetString("repo.path"), &git.CloneOptions{URL: repoUrl})
		if err != nil {
			logrus.Fatalf("error initializing repo:%s", err)
		}
		err = cfgRepo.Init()
		if err != nil {
			logrus.Fatalf("error loading the config repo:%s", err)
		}
		err = cfgRepo.StartPullingEvery(5 * time.Second)
		if err != nil {
			logrus.Fatalf("error pulling the repo:%s", err)
		}

		restSrv := rest.New(cfgRepo, ":8080")
		err = restSrv.Start()
		if err != nil {
			logrus.Fatalf("error starting rest server:%s", err)
		}
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
		viper.SetConfigName("vconf")
		viper.SetConfigType("yaml")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
