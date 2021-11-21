package main

import (
	"context"
	"fmt"
	"github.com/cloudinary/cloudinary-go"
	"github.com/cloudinary/cloudinary-go/api/uploader"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
	"golang.org/x/sync/errgroup"
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

	err = h.container.Run(tmpFile.Name())
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

func (h *Handler) uploadToCloudinary(ctx context.Context, outputDirectory string) error {
	files, err := ioutil.ReadDir(fmt.Sprintf("data/output/%v", outputDirectory))
	if err != nil {
		return err
	}

	if len(files) == 0 {
		return fmt.Errorf("no stems were found for the song in the directory %v", outputDirectory)
	}

	g, ctx := errgroup.WithContext(ctx)

	uploadResults := make([]*uploader.UploadResult, len(files))
	for i, file := range files {
		i, file := i, file // bind loop variables within closure https://golang.org/doc/faq#closures_and_goroutines

		g.Go(func() error {
			uploadFilePath := fmt.Sprintf("data/output/%v/%v", outputDirectory, file.Name())

			rsp, err := h.cloudinary.Upload.Upload(ctx, uploadFilePath, uploader.UploadParams{
				PublicID: fmt.Sprintf("stems/%v/%v", outputDirectory, stripExtensionFromFileBaseName(file.Name())),
			})
			if err != nil {
				return err
			}

			uploadResults[i] = rsp

			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return fmt.Errorf("failed to upload file: %v", err)
	}

	for _, result := range uploadResults {
		fmt.Printf("Upload url: %v\n", result.URL)
	}

	return nil
}


func main() {
	ctx := context.Background()

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error reading .env file")
	}

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

	r := mux.NewRouter()
	r.HandleFunc("/get-stems", handler.getStemsHandler).Methods(http.MethodPost, http.MethodOptions)
	r.Use(PanicRecoveryMiddleware)

	// Todo: Refactor or rename
	server := cors.Default().Handler(r)

	fmt.Printf("💖 Running on 8000 💖\n")
	log.Fatal(http.ListenAndServe(":8000", server))
}
