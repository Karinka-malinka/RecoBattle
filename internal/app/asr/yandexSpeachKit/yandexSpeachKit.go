package yandexspeachkit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/RecoBattle/cmd/config"
	"github.com/RecoBattle/internal/app/asr"
	"github.com/sirupsen/logrus"
)

type Response struct {
	Data string `json:"result"`
}

type ServiceASRYandex struct {
	cnf    config.YandexAsr
	client *http.Client
}

var _ asr.ASR = &ServiceASRYandex{}

func NewYandexASRStore(cnf config.YandexAsr) *ServiceASRYandex {
	return &ServiceASRYandex{
		cnf: cnf,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (ct ServiceASRYandex) TextFromASRModel(data []byte) (string, error) {
	var result Response

	uri := fmt.Sprintf("%v?topic=%v&folderId=%v&lang=ru-RU&format=%v&sampleRateHertz=%v", ct.cnf.YandexAsrUri, "general", ct.cnf.YandexFolderId, ct.cnf.Format, ct.cnf.SampleRateHertz)
	logrus.Infof("Yandex request uri: %v", uri)

	req, err := http.NewRequest("POST", uri, bytes.NewBuffer(data))
	if err != nil {
		logrus.Errorf("error in creating request to Yandex ASR. error: %v", err)
		return "", err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Api-Key %s", ct.cnf.YandexKey))

	response, err := ct.client.Do(req)

	if err != nil {
		logrus.Errorf("error in doing request to Yandex ASR. error: %v", err)
		return "", err
	}

	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		logrus.Errorf("error in reading response body from Yandex ASR. error: %v", err)
		return "", err
	}

	err = json.Unmarshal(responseBody, &result)
	if err != nil {
		logrus.Errorf("error in umarshaling Yandex ASR body. error: %v", err)
		return "", err
	}

	return result.Data, nil

}
