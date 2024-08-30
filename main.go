package main

import (
	"flag"
	"log"
	"os"

	"github.com/alvarofc/mode/api"
	"github.com/alvarofc/mode/storage"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}

	if err := api.InitializeKeys(); err != nil {
		log.Fatalf("Failed to initialize keys: %v", err)
	}

	listenAddr := flag.String("listenaddr", ":8080", "The address to listen on for HTTP requests.")
	flag.Parse()

	pg, err := storage.NewPostgres(
		os.Getenv("DB_HOST"),     // host
		os.Getenv("DB_PORT"),     // port
		os.Getenv("DB_USER"),     // user
		os.Getenv("DB_PASSWORD"), // password
		os.Getenv("DB_NAME"),     // dbname
	)
	if err != nil {
		log.Fatalf("Error creating postgres client: %v", err)
	}
	s3 := storage.NewS3Client(os.Getenv("KEY_ID"), os.Getenv("APP_KEY"), os.Getenv("S3_URL"), os.Getenv("S3_REGION"))

	server := api.NewServer(*listenAddr, pg, &s3)
	log.Println("Server running on port: ", *listenAddr)
	log.Fatal(server.Start())
}
