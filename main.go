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

package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	webflag "github.com/prometheus/exporter-toolkit/web/kingpinflag"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/mm-dict/pearl-exporter/prober"
)

// Namespace defines the common namespace to be used by all metrics.
const namespace = "pearl"

var (
	webConfig     = webflag.AddFlags(kingpin.CommandLine)
	listenAddress = kingpin.Flag("web.listen-address", "The address to listen on for HTTP requests.").Default(":9115").String()
)

func probeHandler(w http.ResponseWriter, r *http.Request, logger log.Logger) {

	probeSuccessGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "probe_success",
		Help:      "Displays whether or not the probe was a success",
	})
	probeDurationGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "probe_duration_seconds",
		Help:      "Returns how long the probe took to complete in seconds",
	})
	probeInfoGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "system_info",
		Help:      "Returns system info for the probed device",
	}, []string{"firmware_version", "firmware_update_availability", "uptime"})
	probeStorageGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "storage",
		Help:      "Returns the current status for the storage devices attached",
	}, []string{"type"})
	probeCpuGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "cpu_info",
		Help:      "Returns information regarding the systems cpu load and temperature",
	}, []string{"type"})
	probeCpuTempGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "cpu_temp",
		Help:      "Current temperature for the CPU",
	})
	probeRecorderGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "recorder_info",
		Help:      "Returns information regarding the configured recorders",
	}, []string{"id"})
	probeChannelsGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "channels_info",
		Help:      "Returns information regarding the configured channels and their publishers",
	}, []string{"id", "status", "type"})
	probeSDIStatusGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "sdi_status",
		Help:      "Returns information regarding the SDI channel, sets the value to the current fps",
	}, []string{"resolution"})
	probeHDMIStatusGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "hdmi_status",
		Help:      "Returns information regarding the HDMI channel, sets the value to the current fps",
	}, []string{"resolution"})
	probeRCAStatusGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "rca_audio_status",
		Help:      "Returns the current audio levels for the RCA/line in  audio input",
	}, []string{"channel", "type"})

	params := r.URL.Query()
	target := params.Get("target")
	user := params.Get("user")
	password := params.Get("password")
	if target == "" {
		http.Error(w, "Target parameter is missing", http.StatusBadRequest)
		return
	}

	level.Info(logger).Log("msg", "Beginning epiphan pearl probe with username "+user+" and password "+password)

	start := time.Now()
	registry := prometheus.NewRegistry()
	registry.MustRegister(probeSuccessGauge)
	registry.MustRegister(probeDurationGauge)
	registry.MustRegister(probeInfoGauge)
	registry.MustRegister(probeStorageGauge)
	registry.MustRegister(probeRecorderGauge)
	registry.MustRegister(probeChannelsGauge)
	registry.MustRegister(probeCpuGauge)
	registry.MustRegister(probeCpuTempGauge)
	registry.MustRegister(probeSDIStatusGauge)
	registry.MustRegister(probeHDMIStatusGauge)
	registry.MustRegister(probeRCAStatusGauge)

	level.Info(logger).Log("msg", "Probing target : "+target)
	firmwareVersion, firmwareVersionError := prober.GetFirmwareVersion(target, user, password)
	systemInfo, systemInfoError := prober.GetSystemInfo(target, user, password)
	storageInfo, storageInfoError := prober.GetStorageInfo(target, user, password)
	channelInfo, channelInfoError := prober.GetChannelInfo(target, user, password)
	updateInfo, updateInfoError := prober.GetFirmwareUpdateAvailability(target, user, password)
	recorderInfo, recorderInfoError := prober.GetRecorderInfo(target, user, password)
	sdiInfo, sdiInfoError := prober.GetSDIStatus(target, user, password)
	hdmiInfo, hdmiInfoError := prober.GetHDMIStatus(target, user, password)
	rcaInfo, rcaInfoError := prober.GetRCAVolumeStatus(target, user, password)
	duration := time.Since(start).Seconds()

	probeDurationGauge.Set(duration)
	if firmwareVersionError != nil {
		probeSuccessGauge.Set(0)
		level.Info(logger).Log("msg", "Probe failed", "duration_seconds", duration)
	} else {
		probeSuccessGauge.Set(1)
		if updateInfoError == nil {
			probeInfoGauge.With(prometheus.Labels{"firmware_version": firmwareVersion.Result, "firmware_update_availability": updateInfo.Result.Status, "uptime": strconv.FormatInt(int64(systemInfo.Result.Uptime), 10)}).Set(1)
		}

		if systemInfoError == nil {
			probeCpuGauge.WithLabelValues("load").Add(float64(systemInfo.Result.CpuLoad))
			probeCpuGauge.WithLabelValues("load_high").Add(float64(prober.Bool2int(systemInfo.Result.CpuLoadHigh)))
			probeCpuTempGauge.Set(float64(systemInfo.Result.Cputemp))
		}

		if storageInfoError == nil {
			probeStorageGauge.WithLabelValues("total").Add(float64(storageInfo.Result.Total))
			probeStorageGauge.WithLabelValues("free").Add(float64(storageInfo.Result.Free))
		}

		if channelInfoError == nil {
			for key := range channelInfo.Result {
				probeChannelsGauge.With(prometheus.Labels{"id": channelInfo.Result[key].Id,
					"status": channelInfo.Result[key].Status.State, "type": "nosignal"}).Set(float64(channelInfo.Result[key].Status.Nosignal))
				probeChannelsGauge.With(prometheus.Labels{"id": channelInfo.Result[key].Id,
					"status": channelInfo.Result[key].Status.State, "type": "bitrate"}).Set(float64(channelInfo.Result[key].Status.Bitrate))
				probeChannelsGauge.With(prometheus.Labels{"id": channelInfo.Result[key].Id,
					"status": channelInfo.Result[key].Status.State, "type": "duration"}).Set(float64(channelInfo.Result[key].Status.Duration))
			}
		}

		if recorderInfoError == nil {
			for key := range recorderInfo.Result {
				if recorderInfo.Result[key].Status.State == "stopped" {
					probeRecorderGauge.With(prometheus.Labels{"id": recorderInfo.Result[key].Id}).Set(0)
				} else {
					probeRecorderGauge.With(prometheus.Labels{"id": recorderInfo.Result[key].Id}).Set(1)
				}
			}
		}

		if sdiInfoError == nil {
			probeSDIStatusGauge.With(prometheus.Labels{"resolution": sdiInfo.Result[0].Status.Video.Resolution}).Set(float64(sdiInfo.Result[0].Status.Video.Actual_fps))
		} else {
			fmt.Println(sdiInfoError)
		}

		if hdmiInfoError == nil {
			probeHDMIStatusGauge.With(prometheus.Labels{"resolution": hdmiInfo.Result[0].Status.Video.Resolution}).Set(float64(hdmiInfo.Result[0].Status.Video.Actual_fps))
		} else {
			fmt.Println(hdmiInfoError)
		}

		if rcaInfoError == nil {
			probeRCAStatusGauge.With(prometheus.Labels{"channel": "left", "type": "peak"}).Set(float64(rcaInfo.Result.Peak[0]))
			probeRCAStatusGauge.With(prometheus.Labels{"channel": "right", "type": "peak"}).Set(float64(rcaInfo.Result.Peak[1]))
			probeRCAStatusGauge.With(prometheus.Labels{"channel": "left", "type": "rms"}).Set(float64(rcaInfo.Result.Rms[0]))
			probeRCAStatusGauge.With(prometheus.Labels{"channel": "right", "type": "rms"}).Set(float64(rcaInfo.Result.Rms[1]))
		} else {
			fmt.Println(rcaInfoError)
		}
		level.Info(logger).Log("msg", "Probe succeeded", "duration_seconds", duration)
	}

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}

func init() {
	prometheus.MustRegister(version.NewCollector("pearl_exporter"))
}

func main() {
	os.Exit(run())
}

func run() int {
	kingpin.CommandLine.UsageWriter(os.Stdout)
	kingpin.Version(version.Print("pearl_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	logger := log.NewLogfmtLogger(os.Stdout)
	logger = level.NewFilter(logger, level.AllowInfo())
	logger = log.With(logger, "caller", log.DefaultCaller)

	level.Info(logger).Log("msg", "Starting pearl_exporter", "version", version.Info())
	level.Info(logger).Log("build_context", version.BuildContext())

	reg := prometheus.NewRegistry()

	reg.MustRegister(collectors.NewBuildInfoCollector())

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/probe", func(w http.ResponseWriter, r *http.Request) {
		probeHandler(w, r, logger)
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html>
    <head><title>Pearl Exporter</title></head>
    <body>
    <h1>Pearl Exporter</h1>
    <p><a href="probe?target=pearl.local">Probe pearl.local for epiphan pearl metrics</a></p>
    <p><a href="metrics">Metrics</a></p>`))
	})

	srv := &http.Server{Addr: *listenAddress}
	srvc := make(chan struct{})
	term := make(chan os.Signal, 1)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)

	go func() {
		level.Info(logger).Log("msg", "Listening on address", "address", *listenAddress)
		if err := web.ListenAndServe(srv, *webConfig, logger); err != http.ErrServerClosed {
			level.Error(logger).Log("msg", "Error starting HTTP server", "err", err)
			close(srvc)
		}
	}()

	for {
		select {
		case <-term:
			level.Info(logger).Log("msg", "Received SIGTERM, exiting gracefully...")
			return 0
		case <-srvc:
			return 1
		}
	}

}
