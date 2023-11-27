package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Quote struct {
	Quote USDBRL `json:"USDBRL"`
}

type USDBRL struct {
	Bid string `json:"bid"`
}

const file string = "./quote.sql"
const create string = `
  CREATE TABLE IF NOT EXISTS quote (
  id INTEGER NOT NULL PRIMARY KEY,
  time DATETIME NOT NULL,
  bid TEXT
  );`

var DB *sql.DB

func main() {
	db, err := sql.Open("sqlite3", file)
	if err != nil {
		fmt.Println(err)
	}

	DB = db
	dbCreate()

	RunServer()
}

func RunServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/cotacao", GetQuoteHandler)
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func GetQuoteHandler(w http.ResponseWriter, r *http.Request) {
	quote, err := GetDolarQuote()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	var usdbrl Quote
	err = json.Unmarshal(quote, &usdbrl)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	err = quotePersist(usdbrl.Quote.Bid)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(usdbrl.Quote.Bid)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
}

func GetDolarQuote() ([]byte, error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func dbCreate() error {
	stmt, err := DB.Prepare(create)
	if err != nil {
		return err
	}
	if _, err := stmt.Exec(); err != nil {
		return err
	}
	return nil
}

func quotePersist(bid string) error {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()

	stmt, err := DB.Prepare("INSERT INTO quote (bid,time) VALUES (?,?)")
	if err != nil {
		log.Println(err)
		return err
	}
	_, err = stmt.ExecContext(ctx, bid, time.Now())
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}
