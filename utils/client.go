// Package utils provides utility functions for the application.
package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type Response struct {
	Code       string `json:"code"`
	Codein     string `json:"codein"`
	Name       string `json:"name"`
	High       string `json:"high"`
	Low        string `json:"low"`
	VarBid     string `json:"varBid"`
	PctChange  string `json:"pctChange"`
	Bid        string `json:"bid"`
	Ask        string `json:"ask"`
	Timestamp  string `json:"timestamp"`
	CreateDate string `json:"create_date"`
}

type Data struct {
	USDBRL Cotacao `json:"USDBRL"`
}

func GetCotacao() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := &http.Client{}
	resp, err := client.Do(request)
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
	var data Data
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(data.USDBRL.Bid)

	file, err := os.Create("cotacao.txt")
	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()

	_, err = file.Write([]byte(fmt.Sprintf("Dólar: %s", data.USDBRL.Bid)))
	if err != nil {
		log.Println(err)
		return
	}
}
