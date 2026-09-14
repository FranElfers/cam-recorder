# Cam Recorder

Cam Recorder is a software tool to record video from a Tapo C200 camera. It runs as an OpenRC service on Alpine Linux.

## System requirements

You must have these items to operate this software:

- Alpine Linux v3.24 x86_64.
- An AMD GPU with VA-API support (for example, AMD Lucienne).
- Go (to compile the software).
- FFmpeg (to record and to transcode video).

## Installation

You can install the software with `make` or with the self-contained installer script.

### Method 1: Install from source code

Do these steps to install from source code:

1. Compile the software.
   ```shell
   make
   ```
2. Install the binary, configuration, and service files.
   ```shell
   sudo make install
   ```

### Method 2: Install with the installer script

You can create a standalone installer script to deploy to another machine:

1. Build the installer package.
   ```shell
   make installer
   ```
2. Copy `cam-recorder-install.sh` to the target machine and run it as root:
   ```shell
   sudo ./cam-recorder-install.sh
   ```

## Configuration

You can change the software settings in the `config.json` file. The file has these parameters:

- **`rtsp_url`**: The URL of the camera RTSP stream.
- **`output_dir`**: The directory to save the MP4 video files.
- **`hls_output_dir`**: The directory to save the temporary HLS stream files.
- **`disk_limit_pct`**: The minimum free disk space percentage. If the free space is less than this value, the software deletes the oldest video file.
- **`video_duration`**: The length of each video file in seconds.
- **`compression_days`**: The number of days before the software compresses a video file to the H.265 format.
- **`retention_days`**: The number of days before the software deletes a video file.
- **`port`**: The network port for the web interface.
- **`hwaccel_device`**: The device path for hardware acceleration (for example, `/dev/dri/renderD128`).

## Operation

Do these steps to operate the service:

1. Start the service.
   ```shell
   sudo rc-service cam-recorder start
   ```
2. Open a web browser.
3. Go to the web interface. Use the IP address of the server and the port from the configuration file.
   `http://<server-ip>:8080/`

## Web interface features

The web interface lets you do these tasks:

- **Watch live video**: Transcodes the RTSP stream to HLS on demand when you view the page.
- **Diagnose issues**: Shows real-time status of the RTSP camera connection, the FFmpeg recording process, disk space, and server uptime.
- **Inspect logs**: Expand the log viewer to see recent FFmpeg output and errors.
- **Restart the recorder**: Restart the FFmpeg recording process directly from the browser.
- **Play recordings**: Select and watch recorded MP4 files in the browser.
- **Download recordings**: Download recorded files to your computer.

## API endpoints

The web server provides these HTTP endpoints:

- `GET /`: The main web page.
- `GET /videos`: JSON list of recorded MP4 files.
- `GET /download/<file>`: Downloads a recorded file.
- `GET /hls/stream.m3u8`: The live HLS video playlist.
- `GET /api/keepalive`: Keeps the live HLS stream running while viewers are active.
- `GET /api/stats`: Basic system statistics (uptime, disk space).
- `GET /api/diagnostics`: Detailed diagnostics (camera connectivity, recorder state, logs).
- `POST /api/recording/restart`: Restarts the FFmpeg recording process.

## Service logs

You can look at the service logs to troubleshoot problems. The system saves the logs in these files:

- Standard messages: `/var/log/cam-recorder.log`
- Error messages: `/var/log/cam-recorder.err`

To see live error messages, use this command:

```shell
tail -f /var/log/cam-recorder.err
```

## Maintenance tasks

The software does these tasks automatically every hour:

- It compresses old video files to use less disk space.
- It deletes video files that are older than the retention limit.
- It deletes the oldest video files if the free disk space is too low.
