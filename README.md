# Pearl Exporter

A Prometheus exporter for Epiphan Pearl devices, written in Python.

This exporter queries the Epiphan Pearl API for various status information (firmware, system, storage, recorders, channels, SDI, HDMI, audio) and exposes them as Prometheus metrics.

## Usage

### Docker

#### Build

To build the Docker image, run:

```bash
docker build -t pearl-exporter .
```

#### Run

To run the Docker container:

```bash
docker run -p 9115:9115 pearl-exporter
```

### Manual

1.  Install dependencies:
    ```bash
    pip install -r requirements.txt
    ```

2.  Run the exporter:
    ```bash
    python3 pearl_exporter.py
    ```

## Credentials (OpenBao)

The exporter looks up each device's credentials in [OpenBao](https://openbao.org/)
(Vault-compatible) keyed by the device hostname, so credentials no longer have to be
passed in the scrape URL.

### Secret layout

All devices live under one common KV **v2** path, with the device hostname as the final
path segment. Each secret has `username` and `password` keys:

```bash
bao kv put secret/pearl-mini/pearl.local username=admin password=secret
```

Here the mount is `secret`, the prefix is `pearl-mini`, and `pearl.local` is the device
hostname (derived from the `target` URL).

### Configuration

Configure the exporter via environment variables (`BAO_*` take precedence, `VAULT_*` are
accepted as fallbacks so Nomad's native integration vars work directly):

| Variable | Default | Description |
|----------|---------|-------------|
| `BAO_ADDR` / `VAULT_ADDR` | — | OpenBao server address |
| `BAO_TOKEN` / `VAULT_TOKEN` | — | Token used to authenticate to OpenBao |
| `BAO_KV_MOUNT` | `secret` | KV v2 mount point |
| `BAO_PATH_PREFIX` | (empty) | Common path under the mount before the hostname |
| `BAO_CACERT` / `VAULT_CACERT` | — | Optional CA bundle for TLS verification |

### Nomad

Let Nomad render the token into the task environment; the exporter reads it from
`VAULT_TOKEN`:

```hcl
vault {}

template {
  data        = "VAULT_TOKEN={{ env \"VAULT_TOKEN\" }}"
  destination = "secrets/bao.env"
  env         = true
}
```

(With Nomad Workload Identity / native integration, the token is provided automatically;
the exporter stays agnostic to how it was obtained.)

## Metrics

The exporter exposes metrics on the `/probe` endpoint. You only need to provide the
`target` query parameter; credentials are fetched from OpenBao.

Example:

```bash
curl "http://localhost:9115/probe?target=http://pearl.local"
```

`user` and `password` query parameters are still accepted as a **fallback** when the
OpenBao lookup yields nothing:

```bash
curl "http://localhost:9115/probe?target=http://pearl.local&user=admin&password=password"
```

### Exposed Metrics

-   `probe_success`: Displays whether or not the probe was a success.
-   `probe_duration_seconds`: Returns how long the probe took to complete in seconds.
-   `pearl_system_info`: Returns system info for the probed device (labels: `firmware_version`, `uptime`).
-   `pearl_storage`: Returns the current status for the storage devices attached (labels: `type`).
-   `pearl_cpu_info`: Returns information regarding the system's CPU load (labels: `type`).
-   `pearl_cpu_temp`: Current temperature for the CPU.
-   `pearl_recorder_info`: Returns information regarding the configured recorders (labels: `id`).
-   `pearl_channels_info`: Returns information regarding the configured channels (labels: `id`, `status`, `type`).
-   `pearl_sdi_status`: Returns information regarding the SDI channel (labels: `resolution`).
-   `pearl_hdmi_status`: Returns information regarding the HDMI channel (labels: `resolution`).
-   `pearl_rca_audio_status`: Returns the current audio levels for the RCA/line-in audio input (labels: `channel`, `type`).
-   `pearl_xlr_audio_status`: Returns the current audio levels for the XLR audio input (labels: `channel`, `type`).
