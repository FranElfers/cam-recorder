package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"
)

type Config struct {
	RTSPURL         string  `json:"rtsp_url"`
	OutputDir       string  `json:"output_dir"`
	HLSOutputDir    string  `json:"hls_output_dir"`
	DiskLimitPct    float64 `json:"disk_limit_pct"`
	VideoDuration   string  `json:"video_duration"` // Used as segment time in seconds (e.g. "3600")
	CompressionDays float64 `json:"compression_days"`
	RetentionDays   float64 `json:"retention_days"`
	Port            string  `json:"port"`
	HWAccelDevice   string  `json:"hwaccel_device"`
}

var (
	config    Config
	startTime time.Time

	hlsMu      sync.Mutex
	hlsCmd     *exec.Cmd
	hlsLastReq time.Time
	hlsRunning bool
)

func main() {
	startTime = time.Now()

	err := loadConfig("config.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	os.MkdirAll(config.OutputDir, 0755)
	os.MkdirAll(config.HLSOutputDir, 0755)

	go recordContinuously()
	go runMaintenance()
	go manageHLSStream()

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/videos", handleVideos)
	http.HandleFunc("/hls/", handleHLS)
	http.HandleFunc("/api/keepalive", handleKeepalive)
	http.HandleFunc("/api/stats", handleStats)
	http.Handle("/download/", http.StripPrefix("/download/", http.FileServer(http.Dir(config.OutputDir))))

	log.Printf("Server listening on port %s", config.Port)
	log.Fatal(http.ListenAndServe(":"+config.Port, nil))
}

func loadConfig(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return json.NewDecoder(file).Decode(&config)
}

func recordContinuously() {
	for {
		log.Println("Starting recording...")
		// Saving to MP4 directly.
		cmd := exec.Command("ffmpeg",
			"-nostdin",
			"-hwaccel", "vaapi",
			"-hwaccel_device", config.HWAccelDevice,
			"-use_wallclock_as_timestamps", "1",
			"-fflags", "+genpts",
			"-rtsp_transport", "tcp",
			"-i", config.RTSPURL,
			"-c:v", "copy",
			"-c:a", "aac",
			"-f", "segment",
			"-segment_time", config.VideoDuration,
			"-reset_timestamps", "1",
			"-strftime", "1",
			filepath.Join(config.OutputDir, "cam-%Y%m%d-%H%M.mp4"),
		)

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err := cmd.Run()
		log.Printf("Recording process ended: %v", err)
		time.Sleep(5 * time.Second) // Wait before restarting
	}
}

func runMaintenance() {
	ticker := time.NewTicker(1 * time.Hour)
	for range ticker.C {
		log.Println("Running maintenance tasks...")
		compressOldFiles()
		deleteOldFiles()
		enforceDiskLimit()
	}
}

func compressOldFiles() {
	cutoff := time.Now().Add(-time.Duration(config.CompressionDays*24) * time.Hour)

	files, err := os.ReadDir(config.OutputDir)
	if err != nil {
		log.Printf("Error reading output dir: %v", err)
		return
	}

	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".mp4") || strings.Contains(f.Name(), "-hevc") {
			continue
		}

		info, err := f.Info()
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoff) {
			inputPath := filepath.Join(config.OutputDir, f.Name())
			outputPath := filepath.Join(config.OutputDir, strings.TrimSuffix(f.Name(), ".mp4")+"-hevc.mp4")

			log.Printf("Compressing %s to H.265 using VA-API...", f.Name())
			cmd := exec.Command("ffmpeg",
				"-hwaccel", "vaapi",
				"-hwaccel_device", config.HWAccelDevice,
				"-hwaccel_output_format", "vaapi",
				"-i", inputPath,
				"-c:v", "hevc_vaapi",
				"-c:a", "copy",
				outputPath,
			)

			if err := cmd.Run(); err == nil {
				os.Remove(inputPath)
				os.Rename(outputPath, inputPath)
				log.Printf("Successfully compressed %s", f.Name())
			} else {
				log.Printf("Failed to compress %s: %v", f.Name(), err)
				os.Remove(outputPath) // Cleanup temp file
			}
		}
	}
}

func deleteOldFiles() {
	cutoff := time.Now().Add(-time.Duration(config.RetentionDays*24) * time.Hour)

	files, err := os.ReadDir(config.OutputDir)
	if err != nil {
		return
	}

	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".mp4") {
			continue
		}
		info, err := f.Info()
		if err == nil && info.ModTime().Before(cutoff) {
			path := filepath.Join(config.OutputDir, f.Name())
			os.Remove(path)
			log.Printf("Deleted old file: %s", path)
		}
	}
}

func enforceDiskLimit() {
	for {
		freePct := getFreeDiskPct(config.OutputDir)
		if freePct > config.DiskLimitPct {
			break
		}

		log.Printf("Disk free space (%.2f%%) <= %.2f%%, deleting oldest file...", freePct, config.DiskLimitPct)
		if !deleteOldestFile() {
			break
		}
	}
}

func deleteOldestFile() bool {
	files, err := os.ReadDir(config.OutputDir)
	if err != nil || len(files) == 0 {
		return false
	}

	type fileInfo struct {
		name string
		mod  time.Time
	}

	var list []fileInfo
	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".mp4") {
			continue
		}
		info, err := f.Info()
		if err == nil {
			list = append(list, fileInfo{f.Name(), info.ModTime()})
		}
	}

	if len(list) == 0 {
		return false
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].mod.Before(list[j].mod)
	})

	oldest := list[0].name
	os.Remove(filepath.Join(config.OutputDir, oldest))
	log.Printf("Deleted oldest file: %s", oldest)
	return true
}

func getFreeDiskPct(path string) float64 {
	var stat syscall.Statfs_t
	syscall.Statfs(path, &stat)
	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bavail * uint64(stat.Bsize)
	if total == 0 {
		return 100.0
	}
	return float64(free) / float64(total) * 100.0
}

func manageHLSStream() {
	for {
		time.Sleep(5 * time.Second)
		hlsMu.Lock()
		if hlsRunning && time.Since(hlsLastReq) > 15*time.Second {
			log.Println("Stopping HLS stream (no active viewers)")
			if hlsCmd != nil && hlsCmd.Process != nil {
				hlsCmd.Process.Kill()
				hlsCmd.Wait()
			}
			hlsRunning = false

			// Cleanup old segments
			files, _ := filepath.Glob(filepath.Join(config.HLSOutputDir, "*"))
			for _, f := range files {
				os.Remove(f)
			}
		}
		hlsMu.Unlock()
	}
}

func handleKeepalive(w http.ResponseWriter, r *http.Request) {
	hlsMu.Lock()
	defer hlsMu.Unlock()

	hlsLastReq = time.Now()
	if !hlsRunning {
		log.Println("Starting HLS stream for viewers...")

		hlsPath := filepath.Join(config.HLSOutputDir, "stream.m3u8")
		hlsRTSP := strings.Replace(config.RTSPURL, "stream1", "stream2", 1)

		// Clean up any old segments before starting
		files, _ := filepath.Glob(filepath.Join(config.HLSOutputDir, "*"))
		for _, f := range files {
			os.Remove(f)
		}

		hlsCmd = exec.Command("ffmpeg",
			"-nostdin",
			"-hwaccel", "vaapi",
			"-hwaccel_device", config.HWAccelDevice,
			"-hwaccel_output_format", "vaapi",
			"-use_wallclock_as_timestamps", "1",
			"-fflags", "+genpts",
			"-rtsp_transport", "tcp",
			"-i", hlsRTSP,
			"-c:v", "h264_vaapi",
			"-g", "15",
			"-b:v", "2M",
			"-c:a", "aac",
			"-f", "hls",
			"-hls_time", "1",
			"-hls_list_size", "3",
			"-hls_flags", "delete_segments",
			hlsPath,
		)

		hlsCmd.Stdout = os.Stdout
		hlsCmd.Stderr = os.Stderr

		if err := hlsCmd.Start(); err != nil {
			log.Printf("Failed to start HLS: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		hlsRunning = true

		go func() {
			err := hlsCmd.Wait()
			log.Printf("HLS process exited: %v", err)
			hlsMu.Lock()
			hlsRunning = false
			hlsMu.Unlock()
		}()
	}

	w.WriteHeader(http.StatusOK)
}

func handleHLS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	http.StripPrefix("/hls/", http.FileServer(http.Dir(config.HLSOutputDir))).ServeHTTP(w, r)
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	http.ServeFile(w, r, "index.html")
}

func handleVideos(w http.ResponseWriter, r *http.Request) {
	files, err := os.ReadDir(config.OutputDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type vidInfo struct {
		Name string    `json:"name"`
		Size float64   `json:"size"`
		Mod  time.Time `json:"-"`
	}

	var list []vidInfo
	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".mp4") {
			continue
		}
		info, err := f.Info()
		if err == nil {
			list = append(list, vidInfo{
				Name: f.Name(),
				Size: float64(info.Size()) / 1024 / 1024,
				Mod:  info.ModTime(),
			})
		}
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].Mod.After(list[j].Mod)
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func getUsedSpaceGB(path string) float64 {
	var size int64
	files, err := os.ReadDir(path)
	if err != nil {
		return 0
	}
	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".mp4") {
			continue
		}
		info, err := f.Info()
		if err == nil {
			size += info.Size()
		}
	}
	return float64(size) / 1024 / 1024 / 1024
}

func handleStats(w http.ResponseWriter, r *http.Request) {
	stats := map[string]interface{}{
		"uptime":         time.Since(startTime).Round(time.Second).String(),
		"free_disk_pct":  getFreeDiskPct(config.OutputDir),
		"retention_days": config.RetentionDays,
		"used_space_gb":  getUsedSpaceGB(config.OutputDir),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
