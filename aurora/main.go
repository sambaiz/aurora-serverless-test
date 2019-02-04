package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	_ "github.com/go-sql-driver/mysql"
)

var dbSrc = fmt.Sprintf(
	"%s:%s@tcp(%s:%s)/%s?parseTime=true",
	os.Getenv("DB_USER"),
	os.Getenv("DB_PASSWORD"),
	os.Getenv("DB_ENDPOINT_ADDRESS"),
	os.Getenv("DB_ENDPOINT_PORT"),
	os.Getenv("DB_DATABASE"),
)

type Response events.APIGatewayProxyResponse

// Handler is our lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context) (Response, error) {

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
