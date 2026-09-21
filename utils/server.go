// Package utils provides utility functions for the application.
package utils

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Cotacao struct {
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

type CotacaoResponse struct {
	USDBRL Cotacao `json:"USDBRL"`
}

func Server() {
	mux := http.NewServeMux()
	mux.HandleFunc("/cotacao", handleRequest)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      handleLog(mux),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
	server.ListenAndServe()
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	// contexto de 200ms para a requisição
	ctx, cancel := context.WithTimeout(r.Context(), 200*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var cotacaoResponse CotacaoResponse
	if err := json.Unmarshal(body, &cotacaoResponse); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	db := createConnection()
	defer db.Close()

	stmt, err := db.Prepare("INSERT INTO cotacao (code, codein, name, high, low, varBid, pctChange, bid, ask, timestamp, create_date) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Println(err)
	}
	defer stmt.Close()

	ctxDatabase, cancel := context.WithTimeout(r.Context(), 10*time.Millisecond)
	defer cancel()

	_, err = stmt.ExecContext(ctxDatabase, cotacaoResponse.USDBRL.Code, cotacaoResponse.USDBRL.Codein, cotacaoResponse.USDBRL.Name, cotacaoResponse.USDBRL.High, cotacaoResponse.USDBRL.Low, cotacaoResponse.USDBRL.VarBid, cotacaoResponse.USDBRL.PctChange, cotacaoResponse.USDBRL.Bid, cotacaoResponse.USDBRL.Ask, cotacaoResponse.USDBRL.Timestamp, cotacaoResponse.USDBRL.CreateDate)
	if err != nil {
		log.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cotacaoResponse)
}

func createConnection() *sql.DB {
	db, err := sql.Open("sqlite3", "file:cotacao.db")
	if err != nil {
		log.Println(err)
	}
	return db
}

func handleLog(next http.Handler) http.Handler {
	serveHTTP := func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.URL.Path, "-", r.Method)
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Println(time.Since(start))
	}

	return http.HandlerFunc(serveHTTP)
}
