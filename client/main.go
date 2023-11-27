package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	RunClient()
}

func RunClient() {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		log.Println(err)
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}
	WriteFile(string(body))
	log.Println("Finished with success")

}

func WriteFile(body string) {
	f, err := os.Create("cotacao.txt")
	if err != nil {
		log.Println(err)
		return
	}
	_, err = f.Write([]byte(fmt.Sprintf("Dólar: %s", body)))
	if err != nil {
		log.Println(err)
		return
	}
}
