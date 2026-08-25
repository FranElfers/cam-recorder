# Cam Recorder

Cam Recorder is a software tool to record video from a Tapo C200 camera. It runs as an OpenRC service on Alpine Linux.

## System requirements

You must have these items to operate this software:
*   Alpine Linux v3.24 x86_64.
*   An AMD GPU with VA-API support (for example, AMD Lucienne).
*   Go (to compile the software).
*   FFmpeg (to record and to transcode video).

## Installation

Do these steps to install the software:

1.  Compile the software.
    ```shell
    go build -o cam-recorder main.go
    ```
2.  Copy the compiled file to the system binaries directory.
    ```shell
    sudo cp cam-recorder /usr/local/bin/
    ```
3.  Make a directory for the configuration file.
    ```shell
    sudo mkdir -p /etc/cam-recorder
    ```
4.  Copy the configuration file to the new directory.
    ```shell
    sudo cp config.json /etc/cam-recorder/
    ```
5.  Copy the OpenRC service file to the system services directory.
    ```shell
    sudo cp cam-recorder.initd /etc/init.d/cam-recorder
    ```
6.  Make the service file executable.
    ```shell
    sudo chmod +x /etc/init.d/cam-recorder
    ```
7.  Add the service to the default runlevel.
    ```shell
    sudo rc-update add cam-recorder default
    ```

## Configuration

You can change the software settings in the `config.json` file. The file has these parameters:

*   **`rtsp_url`**: The URL of the camera RTSP stream.
*   **`output_dir`**: The directory to save the MP4 video files.
*   **`hls_output_dir`**: The directory to save the temporary HLS stream files.
*   **`disk_limit_pct`**: The minimum free disk space percentage. If the free space is less than this value, the software deletes the oldest video file.
*   **`video_duration`**: The length of each video file in seconds.
*   **`compression_days`**: The number of days before the software compresses a video file to the H.265 format.
*   **`retention_days`**: The number of days before the software deletes a video file.
*   **`port`**: The network port for the web interface.
*   **`hwaccel_device`**: The device path for hardware acceleration (for example, `/dev/dri/renderD128`).

## Operation

Do these steps to operate the service:

1.  Start the service.
    ```shell
    sudo rc-service cam-recorder start
    ```
2.  Open a web browser.
3.  Go to the web interface. Use the IP address of the server and the port from the configuration file.
    `http://<server-ip>:8080/`

## Service logs

You can look at the service logs to troubleshoot problems. The system saves the logs in these files:
*   Standard messages: `/var/log/cam-recorder.log`
*   Error messages: `/var/log/cam-recorder.err`

To see the live error messages, use this command:
```shell
tail -f /var/log/cam-recorder.err
```

## Web interface features

The web interface lets you do these tasks:
*   Watch the live video feed. The software transcodes the video to HLS format when you open the page.
*   Look at the system data. You can see the uptime, the free disk space, and the configuration parameters.
*   Watch recorded video files.
*   Download recorded video files to your computer.

## Maintenance tasks

The software does these tasks automatically every hour:
*   It compresses old video files to use less disk space.
*   It deletes video files that are older than the retention limit.
*   It deletes the oldest video files if the free disk space is too low.
