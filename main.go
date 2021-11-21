package main

import (
	"context"
	"fmt"
	"github.com/cloudinary/cloudinary-go"
	"github.com/darthchudi/wav/container"
	"github.com/darthchudi/wav/handler"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
	"log"
	"net/http"
	"os"
)


func main() {
	ctx := context.Background()

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error reading .env file")
	}

	container, err := container.NewContainer(ctx)
	if err != nil {
		log.Fatalf("Failed to create container: %v", err)
	}

	cld, err := cloudinary.NewFromParams(os.Getenv("CLOUDINARY_NAME"), os.Getenv("CLOUDINARY_API_KEY"), os.Getenv("CLOUDINARY_API_SECRET"))
	if err != nil {
		log.Fatalf("Failed to create cloudinary client: %v", err)
	}

	h := &handler.Handler{
		Container: container,
		Cloudinary: cld,
	}

	r := mux.NewRouter()
	r.HandleFunc("/get-stems", h.GetStemsHandler).Methods(http.MethodPost, http.MethodOptions)
	r.Use(handler.PanicRecoveryMiddleware)

	// Todo: Refactor or rename
	server := cors.Default().Handler(r)

	fmt.Printf("💖 Running on 8000 💖\n")
	log.Fatal(http.ListenAndServe(":8000", server))
}
