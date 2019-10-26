package configGitRepo

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

var testBasicPath = fmt.Sprintf("%s/vconf", os.TempDir())

func TestMain(m *testing.M) {
	logrus.SetLevel(logrus.DebugLevel)
	m.Run()
	_ = os.RemoveAll(testBasicPath)
}

func InitRepos(t *testing.T) (string, string) {
	localTmpRepo := fmt.Sprintf("%s/%s", testBasicPath, uuid.New().String())
	remoteTmpRepo := fmt.Sprintf("%s/%s", testBasicPath, uuid.New().String())
	assert.NoError(t, archiver.Unarchive("../../../tests/singleConfigRepo.tgz", remoteTmpRepo))
	return localTmpRepo, remoteTmpRepo + "/singleConfigRepo"
}
