package handler

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	MAX_UPLOAD_SIZE = 32 << 20 // Max upload size is 10MB
	INPUT_DIRECTORY = "data/input"
)

func (h *Handler) GetStemsHandler(w http.ResponseWriter, r *http.Request) {
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

	err = h.Container.Run(tmpFile.Name())
	if err != nil {
		http.Error(w, "failed to get song stems", http.StatusInternalServerError)
		return
	}

	outputDirectory := getOutputDirectoryFromTmpFileName(tmpFile.Name())
	err = h.uploadToCloudinary(r.Context(), outputDirectory)
	if err != nil {
		log.Fatalf("failed to upload song to cdn: %v", err)
		http.Error(w, "failed to upload song to cdn", http.StatusInternalServerError)
		return
	}

	fmt.Printf("💖 Successfully split song in %v\n", time.Since(start))

	w.Write([]byte("Done"))
}

