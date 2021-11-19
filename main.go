package main

import (
	"context"
	"fmt"
	"github.com/cloudinary/cloudinary-go"
	"github.com/cloudinary/cloudinary-go/api/uploader"
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
	cloudinary *cloudinary.Cloudinary
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

	// Validate that song exists
	// Check that folder exists
	err = h.uploadToCloudinary(r.Context(), tmpFile.Name())
	if err != nil {
		log.Fatalf("failed to upload song to cdn: %v", err)
		http.Error(w, "failed to upload song to cdn", http.StatusInternalServerError)
		return
	}

	fmt.Printf("💖 Successfully split song in %v", time.Since(start))

	w.Write([]byte("Done"))
}


// Todo: read env from .env file
// Todo: Get output directory name from file name, comes in format: data/input/169857898-Drake - TSU (Official Audio).mp3
// Todo: Test upload
func (h *Handler) uploadToCloudinary(ctx context.Context, dirname string) error {
	dirname = "169857898-Drake - TSU (Official Audio)"

	files, err := ioutil.ReadDir(fmt.Sprintf("data/output/%v", dirname))
	if err != nil {
		return err
	}

	var uploadResults []*uploader.UploadResult
	for _, file := range files {
		pathToFile := fmt.Sprintf("data/output/%v/%v", dirname, file.Name())
		fmt.Printf("path to file: %v", pathToFile)

		rsp, err := h.cloudinary.Upload.Upload(ctx, pathToFile, uploader.UploadParams{
			PublicID: fmt.Sprintf("stems/%v/%v", dirname, file.Name()),
		})
		if err != nil {
			return err
		}

		uploadResults = append(uploadResults, rsp)
	}

	for _, result := range uploadResults {
		fmt.Printf("Upload url: %v", result.URL)
	}

	return nil
}


func main() {
	ctx := context.Background()

	container, err := NewContainer(ctx)
	if err != nil {
		log.Fatalf("Failed to create container: %v", err)
	}

	cld, err := cloudinary.NewFromParams(os.Getenv("CLOUDINARY_NAME"), os.Getenv("CLOUDINARY_API_KEY"), os.Getenv("CLOUDINARY_API_SECRET"))
	if err != nil {
		log.Fatalf("Failed to create cloudinary client: %v", err)
	}

	handler := &Handler{
		container: container,
		cloudinary: cld,
	}

	log.Fatal(handler.uploadToCloudinary(context.Background(), ""))

	r := mux.NewRouter()
	r.HandleFunc("/get-stems", handler.getStemsHandler).Methods(http.MethodPost, http.MethodOptions)
	r.Use(PanicRecoveryMiddleware)

	// Todo: Refactor or rename
	server := cors.Default().Handler(r)

	fmt.Printf("💖 Running on 8000 💖\n")
	log.Fatal(http.ListenAndServe(":8000", server))
}
