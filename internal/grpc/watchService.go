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
	appRawVer := request.Application.AppVersion
	appVer, err := version.NewVersion(appRawVer)
	if err != nil {
		logrus.Errorf("Error creating version for version %s err:%s", appRawVer, err)
		return err
	}
	watcher := &Watcher{
		id:          uuid.New().String(),
		watcherName: request.WatcherName,
		appName:     request.Application.AppName,
		appVersion:  appVer,
		ch:          make(chan *WatchResponse),
	}
	s.watchers.Store(watcher.id, watcher)

	s.repo.AddOnChangeHandler(func(appName, appVersion string) {
		watcherStreams, err := s.getWatcherStreamByApp(appName, appVersion)
		if err != nil {
			logrus.Errorf("Error getting watcher streams:%s", err)
			return
		}
		for _, watcher := range watcherStreams {
			watcher.ch <- &WatchResponse{Changed: true}
		}
	})
	for {
		select {
		case resp := <-watcher.ch:
			err := stream.Send(resp)
			if err != nil {
				logrus.Errorf("Error sending response:%s", err)
				return err
			}
		case <-stream.Context().Done():
			close(watcher.ch)
			s.watchers.Delete(watcher.id)
			logrus.Infof("watcher %+v removed", watcher)
			return nil
		}
	}
}

func (s *Server) getWatcherStreamByApp(appName, appVersion string) ([]*Watcher, error) {
	cnt, err := version.NewConstraint(fmt.Sprintf("<=%s", appVersion))
	if err != nil {
		logrus.Errorf("Error creating constraint for version %s err:%s", appVersion, err)
		return nil, err
	}

	result := make([]*Watcher, 0)
	s.watchers.Range(func(watcherId, value interface{}) bool {
		watcher := value.(*Watcher)
		if watcher.appName == appName && cnt.Check(watcher.appVersion) {
			result = append(result, watcher)
		}
		return true
	})
	return result, nil
}
