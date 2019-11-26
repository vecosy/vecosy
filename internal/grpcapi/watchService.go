package grpcapi

import (
	"github.com/google/uuid"
	"github.com/hashicorp/go-version"
	"github.com/sirupsen/logrus"
	"github.com/vecosy/vecosy/v2/internal/validation"
	"github.com/vecosy/vecosy/v2/pkg/configrepo"
)

func (s *Server) Watch(request *WatchRequest, stream WatchService_WatchServer) error {
	log := logrus.WithField("method", "Watch").WithField("request", request)
	log.Infof("add Watcher")
	appVersion := configrepo.NewApplicationVersion(request.Application.AppName, request.Application.AppVersion)
	err := validation.ValidateApplicationVersion(appVersion)
	if err != nil {
		log.Errorf("Error validating the application:%+v", appVersion)
		return err
	}
	err = s.CheckToken(stream.Context(), appVersion)
	if err != nil {
		log.Errorf("Error checking token:%s", err)
		return err
	}
	return s.addWatcher(request, stream)
}

func (s *Server) addWatcher(request *WatchRequest, stream WatchService_WatchServer) error {
	appRawVer := request.Application.AppVersion
	appVer, err := validation.ParseVersion(appRawVer)
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

	s.repo.AddOnChangeHandler(func(application configrepo.ApplicationVersion) {
		logrus.Infof("Changes detected on application:%+v", application)
		watcherStreams, err := s.getWatcherStreamByApp(application)
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

func (s *Server) getWatcherStreamByApp(app configrepo.ApplicationVersion) ([]*Watcher, error) {
	newVersion, err := version.NewVersion(app.AppVersion)
	if err != nil {
		logrus.Errorf("Error parsing the application version for version %s err:%s", app.AppVersion, err)
	}
	result := make([]*Watcher, 0)
	s.watchers.Range(func(watcherId, value interface{}) bool {
		watcher := value.(*Watcher)
		if watcher.appName == app.AppName && watcher.appVersion.GreaterThanOrEqual(newVersion) {
			result = append(result, watcher)
		}
		return true
	})
	return result, nil
}
