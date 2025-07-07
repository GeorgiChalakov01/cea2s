package main

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/a-h/templ"
	"github.com/GeorgiChalakov01/cea2s/pages/home"
	"github.com/GeorgiChalakov01/cea2s/pages/part1"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var minioClient *minio.Client
var minioOnce sync.Once
var minioInitErr error

func getMinioClient() (*minio.Client, error) {
	minioOnce.Do(func() {
		endpoint := os.Getenv("MINIO_ENDPOINT")
		accessKey := os.Getenv("MINIO_ACCESS_KEY")
		secretKey := os.Getenv("MINIO_SECRET_KEY")
		
		minioClient, minioInitErr = minio.New(endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
			Secure: false,
		})
	})
	return minioClient, minioInitErr
}

func listAudioFiles(client *minio.Client) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var audioFiles []string
	objectCh := client.ListObjects(ctx, "files", minio.ListObjectsOptions{
		Recursive: true,
	})

	for object := range objectCh {
		if object.Err != nil {
			return nil, object.Err
		}
		if strings.HasSuffix(object.Key, ".mp3") {
			audioFiles = append(audioFiles, object.Key)
		}
	}
	return audioFiles, nil
}

func getRandomAudioFiles(audioFiles []string, count int) []string {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(audioFiles), func(i, j int) {
		audioFiles[i], audioFiles[j] = audioFiles[j], audioFiles[i]
	})
	if len(audioFiles) > count {
		return audioFiles[:count]
	}
	return audioFiles
}

func main() {
	http.Handle("/", http.RedirectHandler("/home", http.StatusSeeOther))

	http.HandleFunc("/home", func(w http.ResponseWriter, r *http.Request) {
		templ.Handler(home.Home()).ServeHTTP(w, r)
	})

	http.HandleFunc("/part1", func(w http.ResponseWriter, r *http.Request) {
		client, err := getMinioClient()
		if err != nil {
			http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
			return
		}

		allAudioFiles, err := listAudioFiles(client)
		if err != nil {
			http.Error(w, "Failed to list audio files", http.StatusInternalServerError)
			return
		}

		randomAudioFiles := getRandomAudioFiles(allAudioFiles, 5)
		templ.Handler(part1.Part1(randomAudioFiles)).ServeHTTP(w, r)
	})

	http.HandleFunc("/audio/", func(w http.ResponseWriter, r *http.Request) {
		client, err := getMinioClient()
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

	fmt.Println("Listening on :8080")
	http.ListenAndServe(":8080", nil)
}
