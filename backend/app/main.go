package main

import (
    "net/http"
    "os"

    "github.com/GeorgiChalakov01/cea2s/handlers"
    "github.com/GeorgiChalakov01/cea2s/lib/db" // New import
    "github.com/GeorgiChalakov01/cea2s/lib/minio"
    "github.com/sirupsen/logrus"
)

func main() {
    // Initialize logger
    logger := logrus.New()
    logger.SetFormatter(&logrus.JSONFormatter{})
    logger.SetOutput(os.Stdout)
    logger.SetLevel(logrus.InfoLevel)

    // Initialize database connection
    if err := db.Connect(); err != nil {
        logger.Fatalf("Failed to connect to database: %v", err)
    }
    defer db.DB.Close()

    // Initialize MinIO service
    minioService, err := minio.GetMinioService(logger)
    if err != nil {
        logger.Fatalf("Failed to initialize MinIO service: %v", err)
    }

    // Setup handlers
    http.Handle("/", http.RedirectHandler("/home", http.StatusSeeOther))
    http.HandleFunc("/home", handlers.HomeHandler)
    http.HandleFunc("/part1", handlers.Part1Handler(minioService))
    http.HandleFunc("/audio/", handlers.AudioHandler(minioService))
    http.HandleFunc("/upload-response", handlers.UploadResponseHandler(minioService, logger))

    // Start server
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    logger.Infof("Server listening on :%s", port)
    if err := http.ListenAndServe(":"+port, nil); err != nil {
        logger.Fatalf("Server failed: %v", err)
    }
}
