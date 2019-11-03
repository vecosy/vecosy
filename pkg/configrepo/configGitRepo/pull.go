package configGitRepo

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"gopkg.in/src-d/go-git.v4"
	"time"
)

func (cr *GitConfigRepo) StartPullingEvery(period time.Duration) error {
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

func (cr *GitConfigRepo) StopPulling() {
	cr.pullCh <- true
}

func (cr *GitConfigRepo) Pull() error {
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
		} else {
			cr.callChangeHandlers()
		}
		return cr.loadApps()
	} else {
		return fmt.Errorf("cannot pull:no remote information found")
	}
}

func (cr *GitConfigRepo) callChangeHandlers() {
	for _, chHandler := range cr.changesHandlers {
		//FIXME : detect the specific appName and appVersion
		chHandler("*","")
	}
}
