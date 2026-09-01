# Repository Analysis

## General Information

This repository contains the code for a camera recorder. The software uses the Go programming language. It records video from a Tapo C200 camera. It runs on Alpine Linux.

## System Architecture

The software runs as a system service. It uses OpenRC to start and stop. It uses FFmpeg to record the video streams.

## Main Components

### Configuration

The software reads settings from a `config.json` file. This file controls the disk limits, the video length, and the network ports.

### Video Recording

A background process records the video stream. It saves the stream to MP4 files. The software uses hardware acceleration to decrease the CPU load.

### File Maintenance

A background process runs every hour. It monitors the free disk space. It deletes old files to increase the free disk space. It compresses old video files to use less disk space.

### Web Server

The software includes a web server. The web server lets the user watch the live video. It lets the user download the recorded files. The web server shows the system data.

## Code Structure

The code is short. It uses the standard Go libraries. This design makes the software easy to read and maintain.
