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

type FirmwareVersion struct {
	Status string
	Result string
}

type SystemStatus struct {
	Status string
	Result SystemStatusDetails
}

type SystemStatusDetails struct {
	Date        string
	Uptime      int64
	CpuLoad     int64
	CpuLoadHigh bool
	Cputemp     int64
}

type FirmwareControl struct {
	Status string
	Result FirmwareControlDetails
}

type FirmwareControlDetails struct {
	Status    string
	Allowed   *string
	Version   *string
	timestamp int64
	Changed   bool
}

type StorageStatus struct {
	Status string
	Result StorageStatusDetails
}

type StorageStatusDetails struct {
	State string
	Total int64
	Free  int64
}

type RecorderStatusDetails struct {
	Id     string
	Status RecorderStatusDetailsRecorderDetails
}

type RecorderStatusDetailsRecorderDetails struct {
	State    string
	Duration *int64
	Active   *string
	Total    *string
}

type RecorderStatus struct {
	Status string
	Result []RecorderStatusDetails
}

type ChannelStatus struct {
	Status string
	Result []ChannelStatusDetails
}

type ChannelStatusDetails struct {
	Id         string
	Status     ChannelStatusDetailsStatus
	Publishers []ChannelStatusDetailsPublishers
}

type ChannelStatusDetailsStatus struct {
	State    string
	Nosignal float64
	Bitrate  float64
	Duration float64
}

type ChannelStatusDetailsPublishers struct {
	Id     string
	Status ChannelStatusDetailsPublishersDetails
}

type ChannelStatusDetailsPublishersDetails struct {
	IsConfigured bool
	Started      bool
	State        string
	Duration     int64
}

type SDIStatus struct {
	Status string
	Result []SDIStatusDetails
}

type SDIStatusDetails struct {
	Id     string
	Name   string
	Status SDIConnectionStatus
}

type SDIConnectionStatus struct {
	Video SDIVideoConnectionStatus
}

type SDIVideoConnectionStatus struct {
	Actual_fps int
	Interlaced bool
	Resolution string
	State      string
	Vrr        int
}

type HDMIStatus struct {
	Status string
	Result []HDMIStatusDetails
}

type HDMIStatusDetails struct {
	Id     string
	Name   string
	Status HDMIConnectionStatus
}

type HDMIConnectionStatus struct {
	Video HDMIVideoConnectionStatus
}

type HDMIVideoConnectionStatus struct {
	Actual_fps int
	Interlaced bool
	Resolution string
	State      string
	Vrr        int
}

type RCAVolumeStatus struct {
	Status string
	Result RCAVolumeDetails
}

type RCAVolumeDetails struct {
	Peak []float64
	Rms  []float64
}
