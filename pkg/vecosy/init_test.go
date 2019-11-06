package vecosy

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetReportCaller(true)
	m.Run()
}

func TestNew(t *testing.T) {
	t.Skip()
	check := assert.New(t)
	cfg := viper.New()
	cl, err := New("localhost:9081", "app1", "v4.0.0", "dev", cfg)
	check.NoError(err)
	check.NotNil(cl)
	check.NoError(cl.WatchChanges())
	t.Logf("config %+v", cfg)
	check.Equal(cfg.GetString("environment"), "dev")
	check.Equal(cfg.GetString("log.level"), "DEBUG")

	for {
		time.Sleep(1 * time.Second)

	}

}
