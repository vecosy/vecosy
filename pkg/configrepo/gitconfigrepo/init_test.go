package gitconfigrepo

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

var testBasicPath = fmt.Sprintf("%s/vecosy_tests", os.TempDir())

func TestMain(m *testing.M) {
	logrus.SetLevel(logrus.DebugLevel)
	retCode := m.Run()
	_ = os.RemoveAll(testBasicPath)
	os.Exit(retCode)
}

func InitRepos(t *testing.T) (string, string) {
	tmpFolder := uuid.New().String()
	localTmpRepo := fmt.Sprintf("%s/%s_local", testBasicPath, tmpFolder)
	remoteTmpRepo := fmt.Sprintf("%s/%s_remote", testBasicPath, tmpFolder)
	assert.NoError(t, archiver.Unarchive("../../../tests/singleConfigRepo.tgz", remoteTmpRepo))
	return localTmpRepo, remoteTmpRepo + "/singleConfigRepo"
}
