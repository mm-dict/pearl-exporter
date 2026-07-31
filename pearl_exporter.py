import time
import logging
import asyncio
from concurrent.futures import ThreadPoolExecutor
from urllib.parse import urlparse
from sanic import Sanic, response
from prometheus_client import CollectorRegistry, Gauge, generate_latest, CONTENT_TYPE_LATEST

import prober
import bao

# Configure logging
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

NAMESPACE = "pearl"

app = Sanic("PearlExporter")

@app.route("/")
async def index(request):
    return response.html("""<html>
    <head><title>Pearl Exporter</title></head>
    <body>
    <h1>Pearl Exporter</h1>
    <p><a href="probe?target=pearl.local">Probe pearl.local for epiphan pearl metrics</a></p>
    <p><a href="metrics">Metrics</a></p>
    </body></html>""")

@app.route("/metrics")
async def metrics(request):
    return response.text(generate_latest(CollectorRegistry()).decode('utf-8'), content_type=CONTENT_TYPE_LATEST)

@app.route("/probe")
async def probe(request):
    target = request.args.get("target")

    if not target:
        return response.text("Target parameter is missing", status=400)

    # Resolve the device hostname from the target URL (handles scheme, port).
    parsed = urlparse(target if "://" in target else f"//{target}")
    hostname = parsed.hostname or target

    # OpenBao first, fall back to URL query params for backwards compatibility.
    creds = bao.get_credentials(hostname)
    if creds:
        user, password = creds
        source = "openbao"
    else:
        user = request.args.get("user")
        password = request.args.get("password")
        source = "url-param"

    logger.info(f"Beginning epiphan pearl probe for {hostname} (credential source: {source})")

    # Run the blocking probe logic in a separate thread
    registry = await asyncio.to_thread(run_probe, target, user, password)
    
    return response.raw(generate_latest(registry), content_type=CONTENT_TYPE_LATEST)

def run_probe(target, user, password):
    registry = CollectorRegistry()
    
    # Define Metrics
    probe_success = Gauge('probe_success', 'Displays whether or not the probe was a success', registry=registry)
    probe_duration = Gauge('probe_duration_seconds', 'Returns how long the probe took to complete in seconds', registry=registry)
    
    probe_info = Gauge('system_info', 'Returns system info for the probed device', ['firmware_version', 'uptime'], namespace=NAMESPACE, registry=registry)
    probe_storage = Gauge('storage', 'Returns the current status for the storage devices attached', ['type'], namespace=NAMESPACE, registry=registry)
    probe_cpu = Gauge('cpu_info', 'Returns information regarding the systems cpu load and temperature', ['type'], namespace=NAMESPACE, registry=registry)
    probe_cpu_temp = Gauge('cpu_temp', 'Current temperature for the CPU', namespace=NAMESPACE, registry=registry)
    probe_recorder = Gauge('recorder_info', 'Returns information regarding the configured recorders', ['id'], namespace=NAMESPACE, registry=registry)
    probe_channels = Gauge('channels_info', 'Returns information regarding the configured channels and their publishers', ['id', 'status', 'type'], namespace=NAMESPACE, registry=registry)
    probe_sdi_status = Gauge('sdi_status', 'Returns information regarding the SDI channel, sets the value to the current fps', ['resolution'], namespace=NAMESPACE, registry=registry)
    probe_hdmi_status = Gauge('hdmi_status', 'Returns information regarding the HDMI channel, sets the value to the current fps', ['resolution'], namespace=NAMESPACE, registry=registry)
    probe_rca_status = Gauge('rca_audio_status', 'Returns the current audio levels for the RCA/line in audio input', ['channel', 'type'], namespace=NAMESPACE, registry=registry)
    probe_xlr_status = Gauge('xlr_audio_status', 'Returns the current audio levels for the XLR audio input', ['channel', 'type'], namespace=NAMESPACE, registry=registry)
    probe_scheduled_events = Gauge('scheduled_events', 'Returns the number of scheduled events', namespace=NAMESPACE, registry=registry)
    probe_last_recording = Gauge('last_recording', 'Returns the time of the last recording', namespace=NAMESPACE, registry=registry)
    probe_finished_events = Gauge('finished_events', 'Returns the number of finished events', namespace=NAMESPACE, registry=registry)

    start_time = time.time()
    logger.info(f"Probing target : {target}")

    # Fetch every endpoint concurrently over one keep-alive session. Each task is
    # independent: a failure records None for that key only, so one bad call no longer
    # suppresses every other metric. probe_success reflects whether all calls succeeded.
    tasks = {
        "firmware": lambda s: prober.get_firmware_version(target, user, password, session=s).get('result', 'unknown'),
        "system": lambda s: prober.get_system_info(target, user, password, session=s).get('result'),
        "storage": lambda s: prober.get_storage_info(target, user, password, session=s).get('result'),
        "channels": lambda s: prober.get_channel_info(target, user, password, session=s).get('result'),
        "recorders": lambda s: prober.get_recorder_info(target, user, password, session=s).get('result'),
        "sources": lambda s: prober.get_sources_status(target, user, password, session=s).get('result'),
        "rca": lambda s: prober.get_rca_volume_status(target, user, password, session=s).get('result'),
        "xlr": lambda s: prober.get_xlr_volume_status(target, user, password, session=s).get('result'),
        "finished": lambda s: prober.get_finished_events(target, user, password, session=s),
        "scheduled": lambda s: prober.get_scheduled_events(target, user, password, session=s),
    }

    data = {}
    success = True
    session = prober.new_session()
    try:
        with ThreadPoolExecutor(max_workers=len(tasks)) as pool:
            futures = {key: pool.submit(fn, session) for key, fn in tasks.items()}
            for key, future in futures.items():
                try:
                    data[key] = future.result()
                except Exception as e:
                    logger.error(f"Error fetching {key}: {e}")
                    data[key] = None
                    success = False
    finally:
        session.close()

    firmware_version = data.get("firmware")
    system_info = data.get("system")
    storage_info = data.get("storage")
    channel_info = data.get("channels")
    recorder_info = data.get("recorders")
    # SDI + HDMI come back in one list; split by source id.
    sources = data.get("sources") or []
    sdi_info = [s for s in sources if s.get('id') == 'D2P0.sdi']
    hdmi_info = [s for s in sources if s.get('id') == 'D2P0.hdmi-a']
    rca_info = data.get("rca")
    xlr_info = data.get("xlr")
    finished_events = data.get("finished")
    scheduled_events = data.get("scheduled")

    try:
        # Update Metrics

        if firmware_version and system_info:
            probe_info.labels(firmware_version=firmware_version, uptime=str(system_info.get('uptime', 0))).set(1)

        if system_info:
            probe_cpu.labels(type="load").inc(float(system_info.get('cpu_load', 0)))
            # Bool2int logic
            load_high = 1 if system_info.get('cpu_load_high') else 0
            probe_cpu.labels(type="load_high").inc(load_high)
            probe_cpu_temp.set(float(system_info.get('cputemp', 0)))

        if storage_info:
            probe_storage.labels(type="total").inc(float(storage_info.get('total', 0)))
            probe_storage.labels(type="free").inc(float(storage_info.get('free', 0)))

        if channel_info:
            for channel in channel_info:
                c_id = channel.get('id')
                status = channel.get('status', {})
                state = status.get('state')
                
                probe_channels.labels(id=c_id, status=state, type="nosignal").set(float(status.get('nosignal', 0)))
                probe_channels.labels(id=c_id, status=state, type="bitrate").set(float(status.get('bitrate', 0)))
                probe_channels.labels(id=c_id, status=state, type="duration").set(float(status.get('duration', 0)))

        if recorder_info:
            for recorder in recorder_info:
                r_id = recorder.get('id')
                status = recorder.get('status', {})
                state = status.get('state')
                val = 0 if state == "stopped" else 1
                probe_recorder.labels(id=r_id).set(val)

        if sdi_info and len(sdi_info) > 0:
            # Go code accesses index 0
            sdi_item = sdi_info[0]
            video_status = sdi_item.get('status', {}).get('video', {})
            probe_sdi_status.labels(resolution=video_status.get('resolution', 'unknown')).set(float(video_status.get('actual_fps', 0)))

        if hdmi_info and len(hdmi_info) > 0:
            hdmi_item = hdmi_info[0]
            video_status = hdmi_item.get('status', {}).get('video', {})
            probe_hdmi_status.labels(resolution=video_status.get('resolution', 'unknown')).set(float(video_status.get('actual_fps', 0)))

        if rca_info:
            peaks = rca_info.get('peak', [])
            rms = rca_info.get('rms', [])
            if peaks and len(peaks) >= 2:
                probe_rca_status.labels(channel="left", type="peak").set(float(peaks[0]))
                probe_rca_status.labels(channel="right", type="peak").set(float(peaks[1]))
            if rms and len(rms) >= 2:
                probe_rca_status.labels(channel="left", type="rms").set(float(rms[0]))
                probe_rca_status.labels(channel="right", type="rms").set(float(rms[1]))

        if xlr_info:
            peaks = xlr_info.get('peak', [])
            rms = xlr_info.get('rms', [])
            if peaks and len(peaks) >= 2:
                probe_xlr_status.labels(channel="left", type="peak").set(float(peaks[0]))
                probe_xlr_status.labels(channel="right", type="peak").set(float(peaks[1]))
            if rms and len(rms) >= 2:
                probe_xlr_status.labels(channel="left", type="rms").set(float(rms[0]))
                probe_xlr_status.labels(channel="right", type="rms").set(float(rms[1]))

        if scheduled_events:
            probe_scheduled_events.set(scheduled_events)

        if finished_events:
            probe_finished_events.set(finished_events.get("number"))
            last_recording = finished_events.get("last_recording")
            if last_recording is not None:
                probe_last_recording.set(last_recording)

        probe_success.set(1 if success else 0)

    except Exception as e:
        logger.error(f"Probe failed with unexpected error: {e}")
        probe_success.set(0)
    
    duration = time.time() - start_time
    probe_duration.set(duration)
    logger.info(f"Probe finished, duration: {duration}")
    
    return registry