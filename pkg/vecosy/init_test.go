package vecosy

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMain(m *testing.M) {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetReportCaller(true)
	m.Run()
}

func TestNew(t *testing.T) {
	cfg := viper.New()
	cl, err := New("localhost:8081", "app1", "1.0.0", "dev", cfg)
	assert.NoError(t, err)
	assert.NotNil(t, cl)
	t.Logf("config %+v", cfg)
	assert.Equal(t, cfg.GetString("environment"), "dev")
	assert.Equal(t, cfg.GetString("log.level"), "DEBUG")
}
