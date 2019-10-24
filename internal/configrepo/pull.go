package configrepo

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/src-d/go-git.v4"
	"time"
)

func (cr *ConfigRepo) StartPullingEvery(period time.Duration) error {
	t := time.NewTicker(period)
	go func() {
		for {
			select {
			case t := <-t.C:
				logrus.Debugf("Auto pull :%+s", t)
				cr.pushError(cr.Pull())
			case <-cr.pullCh:
				t.Stop()
				return
			}
		}
	}()
	return nil
}

func (cr *ConfigRepo) StopPulling() {
	cr.pullCh <- true
}

func (cr *ConfigRepo) Pull() error {
	logrus.Info("Pull")
	if cr.cloneOpts != nil {
		fetchOpts := &git.FetchOptions{Auth: cr.cloneOpts.Auth, Force: true, Tags: git.AllTags}
		err := cr.repo.Fetch(fetchOpts)
		if err != nil {
			if err != git.NoErrAlreadyUpToDate {
				logrus.Errorf("Error pulling :%s", err)
				return err
			} else {
				logrus.Info("already up to date")
			}
		}
		return cr.LoadApps()
	} else {
		logrus.Warn("Cannot pull:no remote information found")
	}
	return nil
}
