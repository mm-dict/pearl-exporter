package prober

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/go-kit/log/level"
)

func getFirmwareVersion(target string, user string, password string, logger log.Logger) FirmwareVersion {
	var f FirmwareVersion
	target = target + "/api/system/firmware/version"
	response := doRequest(target, user, password, logger)
	err := json.Unmarshal(response, &f)
	if err != nil {
		fmt.Println(err)
	}
	return f
}

func getStorageInfo(target string, user string, password string, logger log.Logger) StorageStatus {
	var s StorageStatus
	target = target + "/api/system/storages/main/status"
	response := doRequest(target, user, password, logger)
	err := json.Unmarshal(response, &s)
	if err != nil {
		fmt.Println(err)
	}
	return s
}

func getSystemInfo(target string, user string, password string, logger log.Logger) SystemStatus {
	var s SystemStatus
	target = target + "/api/system/status"
	response := doRequest(target, user, password, logger)
	fmt.Println(response)
	err := json.Unmarshal(response, &s)
	if err != nil {
		fmt.Println(err)
	}
	return s
}

func doRequest(target string, user string, password string, logger log.Logger) []byte {
	client := &http.Client{}

	level.Info(logger).Log("msg", "Probing url : "+target)

	req, err := http.NewRequest("GET", target, nil)
	if err != nil {
		fmt.Println(err)
	}
	req.SetBasicAuth(user, password)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	return bodyBytes
}

func bool2int(b bool) int64 {
	if b {
		return 1
	}
	return 0
}
