package db

import (
    "database/sql"
    "fmt"
    "os"

    _ "github.com/lib/pq"
)

type Part1Question struct {
    ID           int
    QuestionText string
    AudioFile    string
}

var DB *sql.DB

func Connect() error {
    user := os.Getenv("POSTGRES_USER")
    password := os.Getenv("POSTGRES_PASSWORD")
    dbname := os.Getenv("POSTGRES_DB")
    host := "database" // Matches service name in docker-compose

    connStr := fmt.Sprintf("user=%s password=%s dbname=%s host=%s sslmode=disable",
        user, password, dbname, host)
    
    var err error
    DB, err = sql.Open("postgres", connStr)
    if err != nil {
        return err
    }
    
    return DB.Ping()
}

func GetRandomPart1Questions(limit int) ([]Part1Question, error) {
    rows, err := DB.Query(`
        SELECT id, question_text, audio_filename 
        FROM part1_questions 
        ORDER BY RANDOM() 
        LIMIT $1
    `, limit)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var questions []Part1Question
    for rows.Next() {
        var q Part1Question
        if err := rows.Scan(&q.ID, &q.QuestionText, &q.AudioFile); err != nil {
            return nil, err
        }
        questions = append(questions, q)
    }
    return questions, nil
}
