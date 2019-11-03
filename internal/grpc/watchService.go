package vconf

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/go-version"
	"github.com/sirupsen/logrus"
)

func (s *Server) Watch(request *WatchRequest, stream WatchService_WatchServer) error {
	logrus.Infof("add Watcher :%+v", request)
	return s.addWatcher(request, stream)
}

func (s *Server) addWatcher(request *WatchRequest, stream WatchService_WatchServer) error {
	apps := make(map[string]*version.Version)
	appRawVer := request.Application.AppVersion
	appVer, err := version.NewVersion(appRawVer)
	if err != nil {
		logrus.Errorf("Error creating version for version %s err:%s", appRawVer, err)
		return err
	}
	apps[request.Application.AppName] = appVer
	watcher := &Watcher{
		id:          uuid.New().String(),
		watcherName: request.WatcherName,
		apps:        apps,
	}
	s.watchers.Store(watcher.id, watcher)
	s.watcherStreams.Store(watcher.id, stream)

	s.repo.AddOnChangeHandler(func(appName, appVersion string) {
		watcherStreams, err := s.getWatcherStreamByApp(appName, appVersion)
		if err != nil {
			logrus.Errorf("Error getting watcher streams:%s", err)
			return
		}
		for _, watcher := range watcherStreams {
			err := watcher.Send(&WatchResponse{Changed: true})
			if err != nil {
				logrus.Errorf("Error sending watchResponse %s", err)
			}
		}
	})
	return nil
}

func (s *Server) getWatcherStreamByApp(appName, appVersion string) ([]WatchService_WatchServer, error) {
	result := make([]WatchService_WatchServer, 0)
	var mainErr error
	s.watchers.Range(func(watcherId, value interface{}) bool {
		cnt, err := version.NewConstraint(fmt.Sprintf("<=%s", appVersion))
		if err != nil {
			logrus.Errorf("Error creating constraint for version %s err:%s", appVersion, err)
			mainErr = err
			return false
		}
		if cnt.Check(value.(*Watcher).apps[appName]) {
			if stream, found := s.watcherStreams.Load(watcherId); found {
				result = append(result, stream.(WatchService_WatchServer))
			}
		}
		return true
	})
	return result, mainErr
}
