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
	Nosignal int64
	Bitrate  int64
	Duration int64
}

type ChannelStatusDetailsPublishers struct {
	Id     string
	Status ChannelStatusDetailsPublishersDetails
}

type ChannelStatusDetailsPublishersDetails struct {
	IsConfigured bool
	Started      bool
	State        string
	duration     int64
}
