package handler

import (
	"fmt"
	"github.com/cloudinary/cloudinary-go/api/uploader"
	"golang.org/x/sync/errgroup"
	"io/ioutil"
	"context"
)

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

			rsp, err := h.Cloudinary.Upload.Upload(ctx, uploadFilePath, uploader.UploadParams{
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

