package configwatcher

import (
	"context"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"io/ioutil"
)

type configFileWatcher struct {
	filePath string
	context  context.Context
	watcher  *fsnotify.Watcher
	log      *logrus.Entry
}

func NewConfigFileWatcher(filePath string, log *logrus.Entry) (ConfigWatcher, error) {
	return &configFileWatcher{
		filePath: filePath,
		context:  context.Background(),
		log:      log,
	}, nil
}

func (this *configFileWatcher) Get() (string, error) {
	data, err := ioutil.ReadFile(this.filePath)
	if err != nil {
		return "", fmt.Errorf("watcher on '%s': Unable to read '%s': '%v'", this.GetName(), this.filePath, err)
	}
	return string(data), nil
}

func (this *configFileWatcher) GetName() string {
	return fmt.Sprintf("file://%s", this.filePath)
}

func (this *configFileWatcher) Watch(callback func(data string)) error {
	var err error
	this.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("watcher on '%s': Unable to initialize file watcher: '%v'", this.GetName(), err)
	}
	go func() {
		for {
			select {
			case event, ok := <-this.watcher.Events:
				if !ok {
					this.log.Errorf("watcher on '%s': Users config file reload watcher has been closed (1). No more automatic reload!", this.GetName())
					return
				}
				//config.Log.Debugf("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					this.log.Debugf("watcher on '%s': modified file:%s", this.GetName(), event.Name)
					data, err := ioutil.ReadFile(this.filePath)
					if err != nil {
						this.log.Errorf("watcher on '%s': Error on reloading users config file: '%v'. Keep old version", this.GetName(), err)
					}
					callback(string(data))
				}
			case err, ok := <-this.watcher.Errors:
				if !ok {
					this.log.Errorf("watcher on '%s': users config file reload watcher has been closed (2). No more automatic reload!", this.GetName())
					return
				}
				this.log.Errorf("watcher on '%s': Error on users config file reload watcher :%v", this.GetName(), err)
			}
		}
	}()
	err = this.watcher.Add(this.filePath)
	if err != nil {
		return fmt.Errorf("watcher on '%s': Unable to set file watcher on '%s': '%v'", this.GetName(), this.filePath, err)
	}
	return nil
}

func (this *configFileWatcher) Close() {
	if this.watcher != nil {
		_ = this.watcher.Close()
	}
}
