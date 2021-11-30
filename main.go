package main

import (
	"context"
	"github.com/porky256/rest-api/api"
	"github.com/porky256/rest-api/db"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	dbUser, dbPassword, dbName :=
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB")
	dataBase, err := db.Initialize(dbUser, dbPassword, dbName)

	//	dataBase, err := db.Initialize("postgres", "2341", "books")
	if err != nil {
		log.Fatalln(err)
		return
	}
	defer dataBase.Conn.Close()

	handler := api.InitializeHandler(&dataBase)
	server := &http.Server{Addr: ":8080", Handler: handler.Router}
	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Println("listen: ", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = server.Shutdown(ctx); err != nil {
		log.Fatalln("Server error: ", err)
	}
	log.Println("Server shouted down")
}
