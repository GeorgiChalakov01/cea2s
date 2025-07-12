package handlers

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "os"
    "os/exec"
    "path/filepath"
    "time"

    "github.com/GeorgiChalakov01/cea2s/lib/db" // Add this import
    "github.com/GeorgiChalakov01/cea2s/lib/minio"
    "github.com/GeorgiChalakov01/cea2s/pages/home"
    "github.com/GeorgiChalakov01/cea2s/pages/part1"
    "github.com/a-h/templ"
    "github.com/google/uuid"
    "github.com/sirupsen/logrus"
)

// HomeHandler handles home page requests
func HomeHandler(w http.ResponseWriter, r *http.Request) {
    templ.Handler(home.Home()).ServeHTTP(w, r)
}

// Part1Handler handles Part 1 practice page
func Part1Handler(minioService *minio.Service) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        questions, err := db.GetRandomPart1Questions(5) // Now using db package
        if err != nil {
            http.Error(w, "Failed to load questions", http.StatusInternalServerError)
            return
        }

        var audioFiles []string
        for _, q := range questions {
            audioFiles = append(audioFiles, q.AudioFile)
        }

        templ.Handler(part1.Part1(audioFiles)).ServeHTTP(w, r)
    }
}

// AudioHandler serves audio files
func AudioHandler(minioService *minio.Service) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        objectName := r.URL.Path[len("/audio/"):]
        ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
        defer cancel()

        object, err := minioService.GetObject(ctx, "part1-questions", objectName)
        if err != nil {
            http.Error(w, "Audio not found", http.StatusNotFound)
            return
        }
        defer object.Close()

        w.Header().Set("Content-Type", "audio/mpeg")
        if _, err := io.Copy(w, object); err != nil {
            http.Error(w, "Error serving audio", http.StatusInternalServerError)
        }
    }
}

// UploadResponseHandler handles recording uploads
func UploadResponseHandler(minioService *minio.Service, logger *logrus.Logger) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }

        // Parse form with 500MB limit
        if err := r.ParseMultipartForm(500 << 20); err != nil {
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

        // Generate unique ID
        responseID := uuid.New().String()

        // Extract question ID
        questionID := minio.ExtractQuestionID(questionFile)
        if questionID == "" {
            http.Error(w, "Invalid question file", http.StatusBadRequest)
            return
        }

        // Create object name
        objectName := fmt.Sprintf("response-%s-question-%s.mp3", responseID, questionID)

        // Respond immediately
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]string{
            "status":     "accepted",
            "objectName": objectName,
        })

        // Process in background
        go processRecording(minioService, recording, objectName, logger)
    }
}

func processRecording(minioService *minio.Service, recording io.Reader, objectName string, logger *logrus.Logger) {
    // Create temp directory
    tempDir, err := os.MkdirTemp("", "recording-*")
    if err != nil {
        logger.Errorf("Failed to create temp dir: %v", err)
        return
    }
    defer os.RemoveAll(tempDir)

    // Save uploaded file
    webmPath := filepath.Join(tempDir, "recording.webm")
    webmFile, err := os.Create(webmPath)
    if err != nil {
        logger.Errorf("Failed to create temp file: %v", err)
        return
    }

    if _, err := io.Copy(webmFile, recording); err != nil {
        logger.Errorf("Failed to save recording: %v", err)
        webmFile.Close()
        return
    }
    webmFile.Close()

    // Convert to MP3
    mp3Path := filepath.Join(tempDir, "recording.mp3")
    cmd := exec.Command(
        "ffmpeg",
        "-y",
        "-i", webmPath,
        "-c:a", "libmp3lame",
        "-b:a", "192k",
        "-ac", "1",
        "-ar", "44100",
        mp3Path,
    )

    var stderr bytes.Buffer
    cmd.Stderr = &stderr

    if err := cmd.Run(); err != nil {
        logger.Errorf("Conversion failed: %v, %s", err, stderr.String())
        return
    }

    // Open converted file
    mp3File, err := os.Open(mp3Path)
    if err != nil {
        logger.Errorf("Failed to open converted file: %v", err)
        return
    }
    defer mp3File.Close()

    // Get file info
    fileInfo, err := mp3File.Stat()
    if err != nil {
        logger.Errorf("Failed to get file info: %v", err)
        return
    }

    // Upload to MinIO
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := minioService.UploadRecording(ctx, objectName, mp3File, fileInfo.Size()); err != nil {
        logger.Errorf("Failed to upload recording: %v", err)
    } else {
        logger.Infof("Successfully uploaded: %s", objectName)
    }
}
