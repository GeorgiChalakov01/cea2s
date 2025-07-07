package lib

import (
	"context"
	"os"
	"strings"
	"sync"
	"time"
	"io"
	"math/rand"


	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var minioClient *minio.Client
var minioOnce sync.Once
var minioInitErr error

func GetMinioClient() (*minio.Client, error) {
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

func ListAudioFiles(client *minio.Client) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var audioFiles []string
	objectCh := client.ListObjects(ctx, "files", minio.ListObjectsOptions{
		Recursive: true,
		Prefix:	"question-",
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

func ExtractQuestionID(filename string) string {
	if strings.HasPrefix(filename, "question-") && strings.HasSuffix(filename, ".mp3") {
		return strings.TrimPrefix(strings.TrimSuffix(filename, ".mp3"), "question-")
	}
	return ""
}

func UploadRecording(client *minio.Client, objectName string, reader io.Reader, size int64) error {
	_, err := client.PutObject(
		context.Background(),
		"files",
		objectName,
		reader,
		size,
		minio.PutObjectOptions{ContentType: "audio/mpeg"},
	)
	return err
}


func GetRandomAudioFiles(audioFiles []string, count int) []string {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(audioFiles), func(i, j int) {
		audioFiles[i], audioFiles[j] = audioFiles[j], audioFiles[i]
	})
	if len(audioFiles) > count {
		return audioFiles[:count]
	}
	return audioFiles
}
