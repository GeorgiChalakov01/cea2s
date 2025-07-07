package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"io"
	"os"
	"os/exec"
	// "strings"
	// "sync"
	"time"

	"github.com/a-h/templ"
	"github.com/GeorgiChalakov01/cea2s/lib"
	"github.com/GeorgiChalakov01/cea2s/pages/home"
	"github.com/GeorgiChalakov01/cea2s/pages/part1"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)


func main() {
	http.Handle("/", http.RedirectHandler("/home", http.StatusSeeOther))
	http.HandleFunc("/home", func(w http.ResponseWriter, r *http.Request) {
		templ.Handler(home.Home()).ServeHTTP(w, r)
	})

	http.HandleFunc("/part1", func(w http.ResponseWriter, r *http.Request) {
		client, err := lib.GetMinioClient()
		if err != nil {
			http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
			return
		}
		allAudioFiles, err := lib.ListAudioFiles(client)
		if err != nil {
			// Return empty slice instead of error
			fmt.Println("Failed to list audio files:", err)
			allAudioFiles = []string{} // Empty slice
		}
		randomAudioFiles := lib.GetRandomAudioFiles(allAudioFiles, 5)
		templ.Handler(part1.Part1(randomAudioFiles)).ServeHTTP(w, r)
	})

	http.HandleFunc("/audio/", func(w http.ResponseWriter, r *http.Request) {
		client, err := lib.GetMinioClient()
		if err != nil {
			http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
			return
		}
		objectName := r.URL.Path[len("/audio/"):]
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		object, err := client.GetObject(ctx, "files", objectName, minio.GetObjectOptions{})
		if err != nil {
			http.Error(w, "Audio not found", http.StatusNotFound)
			return
		}
		defer object.Close()
		w.Header().Set("Content-Type", "audio/mpeg")
		if _, err := io.Copy(w, object); err != nil {
			http.Error(w, "Error serving audio", http.StatusInternalServerError)
		}
	})

	http.HandleFunc("/upload-response", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
		}

		// Parse multipart form with 500MB limit
		if err := r.ParseMultipartForm(500 << 20); err != nil { // 500 MB
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
		}

		// Get form values
		questionFile := r.FormValue("questionFile")
		recording, _, err := r.FormFile("recording")
		if err != nil {
		http.Error(w, "Invalid recording", http.StatusBadRequest)
		return
		}
		defer recording.Close()

		// Generate unique response ID
		responseID := uuid.New().String()
		
		// Extract question ID
		questionID := lib.ExtractQuestionID(questionFile)
		if questionID == "" {
		http.Error(w, "Invalid question file", http.StatusBadRequest)
		return
		}

		// Create object name - always save as MP3
		objectName := fmt.Sprintf("response-%s-question-%s.mp3", responseID, questionID)

		// Return success immediately
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
		"status":	 "accepted",
		"objectName": objectName,
		})

		// Upload and convert in background goroutine
		go func() {
		// Create temp file for uploaded recording
		tempWebm, err := os.CreateTemp("", "recording-*.webm")
		if err != nil {
			fmt.Printf("Failed to create temp file: %v\n", err)
			return
		}
		defer os.Remove(tempWebm.Name())
		defer tempWebm.Close()

		// Save uploaded file to temp file
		if _, err := io.Copy(tempWebm, recording); err != nil {
			fmt.Printf("Failed to save recording: %v\n", err)
			return
		}

		// Convert to MP3
		tempMp3, err := os.CreateTemp("", "converted-*.mp3")
		if err != nil {
			fmt.Printf("Failed to create temp file: %v\n", err)
			return
		}
		defer os.Remove(tempMp3.Name())
		defer tempMp3.Close()

		cmd := exec.Command(
			"ffmpeg",
			"-y", // Overwrite output file without asking
			"-i", tempWebm.Name(),
			"-c:a", "libmp3lame",
			"-b:a", "192k", // Higher bitrate for better quality
			"-ac", "1", // Convert to mono
			"-ar", "44100", // Set sample rate to 44.1kHz
			"-f", "mp3", // Force MP3 format
			tempMp3.Name(),
		)

		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("Conversion failed: %v, %s\n", err, stderr.String())
			return
		}

		// Upload to MinIO
		client, err := lib.GetMinioClient()
		if err != nil {
			fmt.Printf("Failed to get MinIO client: %v\n", err)
			return
		}

		fileInfo, err := tempMp3.Stat()
		if err != nil {
			fmt.Printf("Failed to get file info: %v\n", err)
			return
		}

		mp3File, err := os.Open(tempMp3.Name())
		if err != nil {
			fmt.Printf("Failed to open converted file: %v\n", err)
			return
		}
		defer mp3File.Close()

		_, err = client.PutObject(
			context.Background(),
			"files",
			objectName,
			mp3File,
			fileInfo.Size(),
			minio.PutObjectOptions{ContentType: "audio/mpeg"},
		)
		if err != nil {
			fmt.Printf("Failed to upload recording: %v\n", err)
		} else {
			fmt.Printf("Successfully uploaded: %s\n", objectName)
		}
		}()
	})

	fmt.Println("Listening on :8080")
	http.ListenAndServe(":8080", nil)
}
