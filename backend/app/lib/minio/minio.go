package minio

import (
    "context"
    "io"
    "os"
    "strings"
    "sync"
    "time"

    "github.com/minio/minio-go/v7"
    "github.com/minio/minio-go/v7/pkg/credentials"
    "github.com/sirupsen/logrus"
)

type Service struct {
    client *minio.Client
    logger *logrus.Logger
}

var (
    instance *Service
    once     sync.Once
)

// GetMinioService returns singleton MinIO service instance
func GetMinioService(logger *logrus.Logger) (*Service, error) {
    var initErr error
    once.Do(func() {
        endpoint := os.Getenv("MINIO_ENDPOINT")
        accessKey := os.Getenv("MINIO_ACCESS_KEY")
        secretKey := os.Getenv("MINIO_SECRET_KEY")

        client, err := minio.New(endpoint, &minio.Options{
            Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
            Secure: false,
        })

        if err != nil {
            initErr = err
            return
        }

        instance = &Service{
            client: client,
            logger: logger,
        }
    })

    return instance, initErr
}

// GetObject retrieves an object from MinIO
func (s *Service) GetObject(ctx context.Context, bucketName, objectName string) (*minio.Object, error) {
    return s.client.GetObject(ctx, bucketName, objectName, minio.GetObjectOptions{})
}

// ListAudioFiles lists all audio files in the bucket
func (s *Service) ListAudioFiles(ctx context.Context) ([]string, error) {
    ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()

    var audioFiles []string
    objectCh := s.client.ListObjects(ctx, "part1-questions", minio.ListObjectsOptions{
        Recursive: true,
        Prefix:    "question-",
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

// ExtractQuestionID extracts question ID from filename
func ExtractQuestionID(filename string) string {
    if strings.HasPrefix(filename, "question-") && strings.HasSuffix(filename, ".mp3") {
        return strings.TrimSuffix(strings.TrimPrefix(filename, "question-"), ".mp3")
    }
    return ""
}

// UploadRecording uploads a recording to MinIO
func (s *Service) UploadRecording(ctx context.Context, objectName string, reader io.Reader, size int64) error {
    _, err := s.client.PutObject(
        ctx,
        "part1-answers",
        objectName,
        reader,
        size,
        minio.PutObjectOptions{ContentType: "audio/mpeg"},
    )
    return err
}
