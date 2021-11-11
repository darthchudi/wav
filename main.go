package main

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"time"
)

func bytesToMb(size int64) float64 {
	return float64(size) / math.Pow(10, 6)
}

type Handler struct {
	container *Container
}

func (h *Handler) getStemsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Print("✨ splitting song with spleeter")

	start := time.Now()

	r.Body = http.MaxBytesReader(w, r.Body, MAX_UPLOAD_SIZE)

	err := r.ParseMultipartForm(MAX_UPLOAD_SIZE)
	if err != nil {
		log.Println("Max file upload reached")
		http.Error(w, fmt.Sprintf("maximum file size is %v", MAX_UPLOAD_SIZE), http.StatusInternalServerError)
		return
	}
	defer r.MultipartForm.RemoveAll()

	formFile, fileHeader, err := r.FormFile("uploadFile")
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to read form file: %v"), http.StatusInternalServerError)
		return
	}
	defer formFile.Close()

	fmt.Printf("File name: %v\n", fileHeader.Filename)
	fmt.Printf("File Size: %vMB\n", bytesToMb(fileHeader.Size))

	tmpFilePattern := fmt.Sprintf("*-%v", fileHeader.Filename)
	tmpFile, err := ioutil.TempFile(INPUT_DIRECTORY, tmpFilePattern)
	if err != nil {
		http.Error(w, "failed to create temp file", http.StatusInternalServerError)
		return
	}
	defer os.Remove(tmpFile.Name()) // clean up temp file

	_, err = io.Copy(tmpFile, formFile)
	if err != nil {
		http.Error(w, "failed to copy uploaded data into temp file", http.StatusInternalServerError)
		return
	}

	err = h.container.Run(tmpFile.Name())
	if err != nil {
		http.Error(w, "failed to get song stems", http.StatusInternalServerError)
		return
	}

	fmt.Printf("💖 Successfully split song in %v", time.Since(start))

	w.Write([]byte("Done"))
}


func main() {
	ctx := context.Background()

	container, err := NewContainer(ctx)
	if err != nil {
		log.Fatalf("Failed to create container: %v", err)
	}

	handler := &Handler{container: container}

	r := mux.NewRouter()
	r.HandleFunc("/get-stems", handler.getStemsHandler).Methods(http.MethodPost, http.MethodOptions)
	r.Use(PanicRecoveryMiddleware)

	// Todo: Refactor or rename
	server := cors.Default().Handler(r)

	fmt.Printf("💖 Running on 8000 💖\n")
	log.Fatal(http.ListenAndServe(":8000", server))
}
