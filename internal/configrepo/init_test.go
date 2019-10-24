package configrepo

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/mholt/archiver"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"os"
	"testing"
	"time"
)

var editorSignature = &object.Signature{
	Name:  "Config Editor",
	Email: "editor@cfg.local",
	When:  time.Now(),
}

func TestMain(m *testing.M) {
	logrus.SetLevel(logrus.DebugLevel)
	m.Run()
}

func InitRepos(t *testing.T) (string, string) {
	localTmpRepo := fmt.Sprintf("%s/vconf/%s", os.TempDir(), uuid.New().String())
	remoteTmpRepo := fmt.Sprintf("%s/vconf/%s", os.TempDir(), uuid.New().String())
	assert.NoError(t, archiver.Unarchive("../../tests/singleConfigRepo.tgz", remoteTmpRepo))
	return localTmpRepo, remoteTmpRepo + "/singleConfigRepo"
}
