package cdn

import (
	"fmt"
	"github.com/muesli/smartcrop"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"

	"code.google.com/p/graphics-go/graphics"
)

func writeByMimetype(w io.Writer, dst image.Image, mimetype string) error {
	switch mimetype {
	case "jpeg":
		return jpeg.Encode(w, dst, &jpeg.Options{jpeg.DefaultQuality})
	case "png":
		return png.Encode(w, dst)
	default:
		return fmt.Errorf("Mimetype '%s' can't be processed.", mimetype)
	}
}

// Smart crop given image file & write it to io.Writer
func smartCrop(w io.Writer, r io.Reader, size []int) error {
	img, mimetype, err := image.Decode(r)
	if size == nil || err != nil {
		io.Copy(w, r)
		return nil
	}

	size = setMaxSize(fitToActualSize(&img, size))
	crop, err := smartcrop.SmartCrop(&img, size[0], size[1])
	if err != nil {
		io.Copy(w, r)
		return nil
	}

	croppedBuffer := image.NewRGBA(image.Rect(0, 0, crop.Width, crop.Height))
	draw.Draw(
		croppedBuffer,
		croppedBuffer.Bounds(),
		img,
		image.Point{crop.X, crop.Y},
		draw.Src,
	)

	dst := image.NewRGBA(image.Rect(0, 0, size[0], size[1]))
	graphics.Scale(dst, croppedBuffer)
	return writeByMimetype(w, dst, mimetype)
}

func fitToActualSize(img *image.Image, size []int) []int {
	ib := (*img).Bounds()
	var x, y int = ib.Dx(), ib.Dy()
	if ib.Dx() >= size[0] {
		x = size[0]
	}
	if ib.Dy() >= size[1] {
		y = size[1]
	}

	return []int{x, y}
}

func setMaxSize(size []int) []int {
	if conf.MaxSize <= size[0] {
		size[0] = conf.MaxSize
	}
	if conf.MaxSize <= size[1] {
		size[1] = conf.MaxSize
	}
	return size
}

// Crop given image file & write it to io.Writer
func crop(w io.Writer, r io.Reader, size []int) error {
	img, mimetype, err := image.Decode(r)
	if size == nil || err != nil {
		io.Copy(w, r)
		return nil
	}

	size = setMaxSize(fitToActualSize(&img, size))
	dst := image.NewRGBA(image.Rect(0, 0, size[0], size[1]))
	graphics.Thumbnail(dst, img)

	return writeByMimetype(w, dst, mimetype)
}

// Resize given image file & write it to io.Writer
func resize(w io.Writer, r io.Reader, size []int) error {
	img, mimetype, err := image.Decode(r)
	if size == nil || err != nil {
		io.Copy(w, r)
		return nil
	}

	ib := img.Bounds()

	size = fitToActualSize(&img, size)
	x := size[0]
	y := size[1]

	// set optimal thumbnail size
	wrat := float64(x) / float64(ib.Dx())
	hrat := float64(y) / float64(ib.Dy())
	if wrat <= hrat {
		y = int(wrat * float64(ib.Dy()))
	} else {
		x = int(hrat * float64(ib.Dx()))
	}

	dst := image.NewRGBA(image.Rect(0, 0, x, y))
	graphics.Thumbnail(dst, img)

	return writeByMimetype(w, dst, mimetype)
}
