import requests
import logging

# Configure logging
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

requests.packages.urllib3.disable_warnings()

def new_session():
    """A keep-alive session for the device's self-signed cert (verify=False)."""
    session = requests.Session()
    session.verify = False  # InsecureSkipVerify: true
    return session

def do_request(url, user, password, method="GET", session=None):
    requester = session if session is not None else requests
    try:
        logger.info(f"Probing url : {url}")
        response = requester.request(
            method,
            url,
            auth=(user, password),
            verify=False,  # InsecureSkipVerify: true
            timeout=5,  # Good practice to add timeout
        )
        response.raise_for_status()
        return response.json()
    except requests.exceptions.RequestException as e:
        logger.error(f"Request failed: {e}")
        raise

def get_firmware_version(target, user, password, session=None):
    url = f"{target}/api/system/firmware/version"
    return do_request(url, user, password, session=session)

def get_storage_info(target, user, password, session=None):
    url = f"{target}/api/system/storages/main/status"
    return do_request(url, user, password, session=session)

def get_system_info(target, user, password, session=None):
    url = f"{target}/api/system/status"
    return do_request(url, user, password, session=session)

def get_recorder_info(target, user, password, session=None):
    url = f"{target}/api/recorders/status"
    return do_request(url, user, password, session=session)

def get_channel_info(target, user, password, session=None):
    url = f"{target}/api/channels/status?publishers=true"
    return do_request(url, user, password, session=session)

def get_sources_status(target, user, password, session=None):
    """SDI + HDMI in a single call; caller splits the result list by source id."""
    url = f"{target}/api/sources/status?ids=D2P0.hdmi-a,D2P0.sdi"
    return do_request(url, user, password, session=session)

def get_rca_volume_status(target, user, password, session=None):
    url = f"{target}/api/sources/D2P0.analog-a/audiolevels"
    return do_request(url, user, password, session=session)

def get_xlr_volume_status(target, user, password, session=None):
    url = f"{target}/api/sources/D2P0.analog-b/audiolevels"
    return do_request(url, user, password, session=session)

def get_finished_events(target, user, password, session=None):
    url = f"{target}/api/schedule/events?status=finished"
    result = do_request(url, user, password, session=session)
    events = result.get("result") or []
    finished_recordings = len(events)
    last_recording = events[-1].get("start") if events else None

    return {
            "number": finished_recordings,
            "last_recording": last_recording
    }


def get_scheduled_events(target, user, password, session=None):
    url = f"{target}/api/schedule/events?status=scheduled"
    result = do_request(url, user, password, session=session)
    scheduled_recordings = len(result.get("result") or [])

    return scheduled_recordings
