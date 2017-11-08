package netWorking

import (
	"Clans/server/log"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	JsonContentType = "application/json"
)

const (
	CreateRoom = "createRoom"
)

type HttpCallBack func(map[string]*json.RawMessage)

func HttpPost(url string, port int, id string, data map[string]interface{}, callBack HttpCallBack) {
	api := fmt.Sprintf("http://%s:%d", url, port)

	dataBytes, _ := json.Marshal(data)

	sendData := map[string]interface{}{
		"ID":      id,
		"RawJson": string(dataBytes),
	}

	raw, err := json.Marshal(sendData)
	if err != nil {
		log.Logger().Errorf("Error when marshal data:%v, err:%v", data, err)
		return
	}
	buf := bytes.NewBuffer(raw)

	req, err := http.NewRequest("POST", api, buf)
	req.Header.Set("Content-Type", JsonContentType)

	client := &http.Client{}
	resp, er := client.Do(req)
	if er != nil {
		log.Logger().Errorf("Error when post data:%v, err:%v", data, er)
		return
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	retDataMap := make(map[string]*json.RawMessage)
	json.Unmarshal(body, &retDataMap)

	callBack(retDataMap)
}
