package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	_ "github.com/go-sql-driver/mysql"
	"github.com/sambaiz/aurora-serverless-test/secret"
)

type dbConfig struct {
	Password string `json:"password"`
	DbName   string `json:"dbname"`
	Engine   string `json:"engine"`
	Port     int    `json:"port"`
	Host     string `json:"host"`
	UserName string `json:"username"`
}

func dbSrc() (string, error) {
	secret, err := secret.GetSecretString(os.Getenv("DB_SECRET"))
	if err != nil {
		return "", err
	}
	var config dbConfig
	if err := json.Unmarshal([]byte(secret), &config); err != nil {
		return "", err
	}
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?parseTime=true",
		config.UserName,
		config.Password,
		config.Host,
		config.Port,
		config.DbName,
	), nil
}

type Response events.APIGatewayProxyResponse

// Handler is our lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context) (Response, error) {
	dbSrc, err := dbSrc()
	if err != nil {
		return Response{StatusCode: http.StatusInternalServerError}, err
	}
	log.Printf("dbSrc: %s", dbSrc)
	db, err := sql.Open("mysql", dbSrc)
	if err != nil {
		return Response{StatusCode: http.StatusInternalServerError}, err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return Response{StatusCode: http.StatusInternalServerError}, err
	}

	_, err = tx.Exec("INSERT INTO b (a_id) VALUES (1)")
	if err != nil {
		tx.Rollback()
		return Response{StatusCode: http.StatusInternalServerError}, err
	}

	rows, err := db.Query("SELECT b.id FROM a JOIN b ON (a.id = b.a_id)")
	if err != nil {
		return Response{StatusCode: http.StatusInternalServerError}, err
	}

	if err := tx.Commit(); err != nil {
		return Response{StatusCode: http.StatusInternalServerError}, err
	}

	bIDs := make([]int64, 0)
	for rows.Next() {
		var bID int64
		if err := rows.Scan(&bID); err != nil {
			return Response{StatusCode: http.StatusInternalServerError}, err
		}
		bIDs = append(bIDs, bID)
	}

	var buf bytes.Buffer

	body, err := json.Marshal(map[string]interface{}{
		"b_ids": bIDs,
	})
	if err != nil {
		return Response{StatusCode: http.StatusInternalServerError}, err
	}
	json.HTMLEscape(&buf, body)

	resp := Response{
		StatusCode:      200,
		IsBase64Encoded: false,
		Body:            buf.String(),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}

	return resp, nil
}

func main() {
	lambda.Start(Handler)
}
