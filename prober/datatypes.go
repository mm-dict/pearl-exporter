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
	duration     int64
}
