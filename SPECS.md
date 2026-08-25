# cam-recorder

System Specifications: Tapo C200 Recording Software

## 1. System Environment
* Target OS: Alpine Linux v3.24 x86_64.
* Execution: Run the program as an OpenRC service.
* Primary Language: Go.
* External Tool: FFmpeg.

## 2. Configuration
* Use a `config.json` file for all settings.
* Include parameters for RTSP URL, disk limits, video duration, and retention times.

## 3. Video Recording
* Record the camera RTSP stream continuously.
* Save the video and audio in 1-hour MP4 files.
* File name syntax: "cam-YYYYMMDD-HHmm.mp4".
* Use `os/exec` to operate FFmpeg for recording.
* Use hardware acceleration for FFmpeg operations. The system GPU is "AMD Lucienne [Integrated]".

## 4. File Management
* Use the standard Go `time.Ticker` for scheduled tasks.
* Compress video files that are older than 48 hours. Use the H.265 codec and AMD hardware acceleration (VA-API/AMF).
* Delete oldest files when the free disk space is 10% or less.
* Delete files that are older than 7 days.

## 5. Web Interface
* Use the `net/http` standard library for the web server.
* Do not use authentication. The interface is open to the local network.
* Transcode the RTSP stream to HLS format for the live video feed using AMD hardware acceleration.
* Process the HLS stream only on-demand (when a user opens the web interface).
* Automatically delete old `.ts` segments to free space.
* Show a list of recorded MP4 files. Include a video player and a download button for the selected file.
* Show system statistics (parameter values, uptime, and free disk space).

## 6. Code Optimization
* Write short and simple code.
* Use standard libraries to decrease the token count.
* Write README.md using Google Docs Style, following Zinsser's four principles of quality writing, and ASD-STE100 Simplified Technical English.
