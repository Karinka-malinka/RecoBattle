package asr

import (
	"sync"

	"github.com/sirupsen/logrus"
)

type ASR interface {
	RegisterASR() string
	TextFromASRModel(data []byte) (string, error)
}

type ASRRegistry struct {
	Services map[string]ASR
	sync.RWMutex
}

func (asrRegistry *ASRRegistry) AddService(name string, service ASR) {

	asrRegistry.Lock()
	defer asrRegistry.Unlock()

	_, exists := asrRegistry.Services[name]
	if exists {
		logrus.Infof("Service [%s] already exists, skipping...\n", name)
		return
	}
	asrRegistry.Services[name] = service
	logrus.Infof("Service [%s] registered.\n", name)
}

func (asrRegistry *ASRRegistry) GetService(name string) (ASR, bool) {
	asrRegistry.RLock()
	defer asrRegistry.RUnlock()
	service, ok := asrRegistry.Services[name]
	return service, ok
}
