package main

import (
	"io"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"regexp"
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
)

var (
	configFile    = kingpin.Flag("config.file", "Pearl exporter configuration file.").Default("pearl.yml").String()
	webConfig     = webflag.AddFlags(kingpin.CommandLine)
	listenAddress = kingpin.Flag("web.listen-address", "The address to listen on for HTTP requests.").Default(":9115").String()
)

func probeHandler(w http.ResponseWriter, r *http.Request, logger log.Logger) {

	probeSuccessGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "probe_success",
		Help: "Displays whether or not the probe was a success",
	})
	probeDurationGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "probe_duration_seconds",
		Help: "Returns how long the probe took to complete in seconds",
	})
	probeInfoGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "probe_system_info",
		Help: "Returns system info for the probed device",
	}, []string{"firmware_version", "uptime"})
	probeStorageGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "probe_storage",
		Help: "Returns the current status for the storage devices attached",
	}, []string{"type"})
	probeCpuGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "probe_cpu_info",
		Help: "Returns information regarding the systems cpu load and temperature",
	}, []string{"type"})
	probeRecordingGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "probe_recording",
		Help: "Returns whether the probed device is currently recording",
	})
	probeStreamingGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "probe_streaming",
		Help: "Returns whether the probed device is currently streaming",
	})

	params := r.URL.Query()
	target := params.Get("target")
	channel := params.Get("channel")
	user := params.Get("user")
	password := params.Get("password")
	if target == "" {
		http.Error(w, "Target parameter is missing", http.StatusBadRequest)
		return
	}

	level.Info(logger).Log("msg", "Beginning epiphan pearl probe")

	start := time.Now()
	registry := prometheus.NewRegistry()
	registry.MustRegister(probeSuccessGauge)
	registry.MustRegister(probeDurationGauge)
	registry.MustRegister(probeInfoGauge)
	registry.MustRegister(probeStorageGauge)
	registry.MustRegister(probeRecordingGauge)
	registry.MustRegister(probeStreamingGauge)
	registry.MustRegister(probeCpuGauge)

	client := &http.Client{}

	url := target + "/admin/" + channel + "/get_params.cgi?rec_enabled&bcast_disabled"
	level.Info(logger).Log("msg", "Probing url : "+url)

	req, err := http.NewRequest("GET", url, nil)
	req.SetBasicAuth(user, password)
	resp, err := client.Do(req)
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	bodyString := string(bodyBytes)

	values := parseResponse(bodyString, logger)
	duration := time.Since(start).Seconds()

	probeDurationGauge.Set(duration)
	if err != nil {
		level.Error(logger).Log("msg", "Probe failed", "duration_seconds", duration)
	} else {
		probeSuccessGauge.Set(1)
		firmwareVersion := exporter.getFirmwareVersion(target, user, password, logger)
		systemInfo := exporter.getSystemInfo(target, user, password, logger)
		probeInfoGauge.With(prometheus.Labels{"firmware_version": firmwareVersion.Result, "uptime": strconv.FormatInt(int64(systemInfo.Result.Uptime), 10)}).Set(1)
		probeCpuGauge.WithLabelValues("load").Add(float64(systemInfo.Result.CpuLoad))
		probeCpuGauge.WithLabelValues("load_high").Add(float64(bool2int(systemInfo.Result.CpuLoadHigh)))
		probeCpuGauge.WithLabelValues("temp").Add(float64(systemInfo.Result.Cputemp))
		storageInfo := exporter.getStorageInfo(target, user, password, logger)
		probeStorageGauge.WithLabelValues("total").Add(float64(storageInfo.Result.Total))
		probeStorageGauge.WithLabelValues("free").Add(float64(storageInfo.Result.Free))

		if values["rec_enabled"] == "on" {
			probeRecordingGauge.Set(1)
		} else {
			probeRecordingGauge.Set(0)
		}
		if values["bcast_disabled"] == "true" {
			probeStreamingGauge.Set(0)
		} else {
			probeStreamingGauge.Set(1)
		}
		level.Info(logger).Log("msg", "Probe succeeded", "duration_seconds", duration)
	}

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}

func parseResponse(resp string, logger log.Logger) map[string]string {
	level.Info(logger).Log("msg", "parsing "+resp)
	//values := strings.split(resp, "\n")
	var rex = regexp.MustCompile("([^=;\r\n]+) = ([^;\r\n]*)")
	data := rex.FindAllStringSubmatch(resp, -1)

	res := make(map[string]string)
	for _, kv := range data {
		level.Info(logger).Log("msg", "key "+kv[1])
		level.Info(logger).Log("msg", "value "+kv[2])
		k := kv[1]
		v := kv[2]
		res[k] = v
	}
	return res
}

func probePearl(target string, registry *prometheus.Registry) bool {
	return true
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
