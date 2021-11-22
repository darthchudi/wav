package handler

import (
	"github.com/cloudinary/cloudinary-go"
	"github.com/darthchudi/wav/container"
)

type Handler struct {
	Container *container.Container
	Cloudinary *cloudinary.Cloudinary
}
