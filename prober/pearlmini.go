// MIT License

// Copyright (c) 2022 Kristof Keppens <kristof.keppens@ugent.be>

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package prober

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

func GetFirmwareVersion(target string, user string, password string) FirmwareVersion {
	var f FirmwareVersion
	target = target + "/api/system/firmware/version"
	response := doRequest(target, user, password, "GET")
	err := json.Unmarshal(response, &f)
	if err != nil {
		fmt.Println(err)
	}
	return f
}

func GetFirmwareUpdateAvailability(target string, user string, password string) FirmwareControl {
	var f FirmwareControl
	target = target + "/api/system/firmware/update/control/check"
	response := doRequest(target, user, password, "POST")
	err := json.Unmarshal(response, &f)
	if err != nil {
		fmt.Println(err)
	}
	return f
}

func GetStorageInfo(target string, user string, password string) StorageStatus {
	var s StorageStatus
	target = target + "/api/system/storages/main/status"
	response := doRequest(target, user, password, "GET")
	err := json.Unmarshal(response, &s)
	if err != nil {
		fmt.Println(err)
	}
	return s
}

func GetSystemInfo(target string, user string, password string) SystemStatus {
	var s SystemStatus
	target = target + "/api/system/status"
	response := doRequest(target, user, password, "GET")
	fmt.Println(response)
	err := json.Unmarshal(response, &s)
	if err != nil {
		fmt.Println(err)
	}
	return s
}

func GetRecorderInfo(target string, user string, password string) RecorderStatus {
	var r RecorderStatus
	target = target + "/api/recorders/status"
	response := doRequest(target, user, password, "GET")
	err := json.Unmarshal(response, &r)
	if err != nil {
		fmt.Println(err)
	}
	return r
}

func GetChannelInfo(target string, user string, password string) ChannelStatus {
	var c ChannelStatus
	target = target + "/api/channels/status?publishers=true"
	response := doRequest(target, user, password, "GET")
	err := json.Unmarshal(response, &c)
	if err != nil {
		fmt.Println(err)
	}
	return c
}

func doRequest(target string, user string, password string, method string) []byte {
	client := &http.Client{}
	logger := log.NewLogfmtLogger(os.Stdout)
	logger = level.NewFilter(logger, level.AllowInfo())
	logger = log.With(logger, "caller", log.DefaultCaller)

	level.Info(logger).Log("msg", "Probing url : "+target)

	req, err := http.NewRequest(method, target, nil)
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

func Bool2int(b bool) int64 {
	if b {
		return 1
	}
	return 0
}
