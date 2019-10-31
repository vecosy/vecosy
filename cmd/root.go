package cmd

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vecosy/vecosy/v2/internal/rest"
	"github.com/vecosy/vecosy/v2/pkg/configrepo/configGitRepo"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
	"path"
	"path/filepath"
	"strings"
)

var cfgFile string
var verboseFlag *bool

var rootCmd = &cobra.Command{
	Use:   "vecosy",
	Short: "VeCoSy - Versioned Configuration System Server",
	Run: func(cmd *cobra.Command, args []string) {
		repoUrl := viper.GetString("repo.remote.url")
		if strings.Contains(repoUrl, "file://") {
			repoPath := strings.Replace(repoUrl, "file://", "",1)
			if !path.IsAbs(repoPath) {
				repoUrl, _ = filepath.Abs(repoPath)
			}
		}
		auth, err := getAuth()
		if err != nil {
			logrus.Fatalf("error initializing repo auth:%s", err)
		}
		cfgRepo, err := configGitRepo.NewConfigRepo(viper.GetString("repo.local.path"), &git.CloneOptions{URL: repoUrl, Auth: auth})
		if err != nil {
			logrus.Fatalf("error initializing repo:%s", err)
		}
		err = cfgRepo.Init()
		if err != nil {
			logrus.Fatalf("error loading the config repo:%s", err)
		}

		pullEvery := viper.GetDuration("repo.remote.pullEvery")
		logrus.Infof("Pull repo every :%s", pullEvery)
		err = cfgRepo.StartPullingEvery(pullEvery)
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

func getAuth() (transport.AuthMethod, error) {
	authType := viper.GetString("repo.remote.auth.type")
	username := viper.GetString("repo.remote.auth.username")
	switch authType {
	case "pubKey":
		keyFile := viper.GetString("repo.remote.auth.keyFile")
		keyFilePassword := viper.GetString("repo.remote.auth.keyFilePassword")
		return ssh.NewPublicKeysFromFile(username, keyFile, keyFilePassword)
	case "plain":
		password := viper.GetString("repo.remote.auth.password")
		return &ssh.Password{
			User:                  username,
			Password:              password,
			HostKeyCallbackHelper: ssh.HostKeyCallbackHelper{},
		}, nil
	case "http":
		password := viper.GetString("repo.remote.auth.password")
		return &http.BasicAuth{
			Username: username,
			Password: password,
		}, nil
	default:
		return nil, nil
	}
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
