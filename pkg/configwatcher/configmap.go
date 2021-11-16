package configwatcher

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"os"
	"path/filepath"
	"reflect"
	"time"
)

type ConfigMapWatcher struct {
	clientSet     *kubernetes.Clientset
	namespace     string
	configMapName string
	configMapKey  string
	context       context.Context
	watcher       watch.Interface
	log           *logrus.Entry
}

func NewConfigMapWatcher(clientSet *kubernetes.Clientset, namespace string, configMapName string, configMapKey string, log *logrus.Entry) (ConfigWatcher, error) {
	var err error
	// We don't handle NamepaceAll, as not relevant in our use case
	if namespace == "" {
		file, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
		if err != nil {
			return nil, fmt.Errorf("watcher on 'cm:???/%s/%s': Unable to lookup current namespace", configMapName, configMapKey)
		}
		namespace = string(file)
	}
	if clientSet == nil {
		clientSet, err = GetClientSet("")
		if err != nil {
			return nil, fmt.Errorf("watcher on 'cm:/%s/%s/%s': Unable to get initialize Kubernetes clientSet: '%v'", namespace, configMapName, configMapKey, err)
		}
	}
	return &ConfigMapWatcher{
		clientSet:     clientSet,
		namespace:     namespace,
		configMapName: configMapName,
		configMapKey:  configMapKey,
		context:       context.Background(),
		log:           log,
	}, nil
}

func (this *ConfigMapWatcher) Get() (string, error) {
	configMap, err := this.clientSet.CoreV1().ConfigMaps(this.namespace).Get(this.context, this.configMapName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("watcher on '%s': Unable to get configmap '%s/%s': '%v'", this.GetName(), this.namespace, this.configMapName, err)
	}
	data, ok := configMap.Data[this.configMapKey]
	if !ok {
		return "", fmt.Errorf("watcher on '%s': Unable to find key '%s' in configMap  '%s/%s'", this.GetName(), this.configMapKey, this.namespace, this.configMapName)
	}
	return data, nil
}

func (this *ConfigMapWatcher) GetName() string {
	return fmt.Sprintf("cm:%s/%s/%s", this.namespace, this.configMapName, this.configMapKey)
}

func (this *ConfigMapWatcher) Watch(callback func(data string)) error {
	var err error
	this.watcher, err = this.clientSet.CoreV1().ConfigMaps(this.namespace).Watch(this.context, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("watcher on '%s': Unable to initialize watcher on configMap: '%v'", this.GetName(), err)
	}
	go func() {
		for {
			select {
			case event, ok := <-this.watcher.ResultChan():
				if !ok {
					this.log.Errorf("watcher on '%s' stopped. Will  restart...", this.GetName())
					for {
						this.watcher, err = this.clientSet.CoreV1().ConfigMaps(this.namespace).Watch(this.context, metav1.ListOptions{})
						if err != nil {
							this.log.Errorf("watcher on '%s': Unable to restart for now. Will retry in 10sec", this.GetName())
							time.Sleep(10 * time.Second)
						} else {
							this.log.Infof("watcher on '%s' restarted successfully", this.GetName())
							break
						}
					}
				}
				configMap, ok := event.Object.(*v1.ConfigMap)
				if !ok {
					this.log.Errorf("watcher on '%s': unexpected type: '%s'", this.GetName(), reflect.TypeOf(event.Object))
					continue
				}
				this.log.Debugf("watcher on '%s': ConfigMap: %s Event type: %v", this.GetName(), configMap.Name, event.Type)
				if configMap.Name == this.configMapName {
					data := ""
					if event.Type == "DELETED" {
						this.log.Warningf("watcher on '%s': ConfigMap '%s/%s' has been deleted", this.GetName(), this.namespace, this.configMapName)
					} else {
						// We handle MODIFIED and also ADDED, even if redundant with Get(), as there is a (very small) risk to loose a modification occuring between Get() and Watch()
						data, ok = configMap.Data[this.configMapKey]
						if !ok {
							this.log.Errorf("watcher on '%s': Unable to find key '%s' in configMap  '%s/%s'", this.GetName(), this.configMapKey, this.namespace, this.configMapName)
						}
					}
					callback(data)
				}
			}
		}
	}()
	return nil
}

func (this *ConfigMapWatcher) Close() {
	if this.watcher != nil {
		this.watcher.Stop()
	}
}

func GetClientSet(kubeconfig string) (*kubernetes.Clientset, error) {
	var config *rest.Config = nil
	var err error
	if kubeconfig == "" {
		if envvar := os.Getenv("KUBECONFIG"); len(envvar) > 0 {
			kubeconfig = envvar
		}
	}
	if kubeconfig == "" {
		config, err = rest.InClusterConfig()
	}
	if config == nil {
		home := homedir.HomeDir()
		if kubeconfig == "" && home != "" {
			kubeconfig = filepath.Join(home, ".kube", "config")
		}
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}
