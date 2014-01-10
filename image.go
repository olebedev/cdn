package main

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"

	"code.google.com/p/graphics-go/graphics"
)

// *** Crop given image file & write it to io.Writer
func crop(w io.Writer, r io.Reader, size []int) error {
	if size == nil {
		io.Copy(w, r)
		return nil
	}

	img, mimetype, err := image.Decode(r)
	if err != nil {
		io.Copy(w, r)
		return err
	}

	ib := img.Bounds()
	// fit to actual size
	var x, y int = ib.Dx(), ib.Dy()
	if ib.Dx() >= size[0] {
		x = size[0]
	}
	if ib.Dy() >= size[1] {
		y = size[1]
	}

	// set max size
	if conf.MaxSize <= x {
		x = conf.MaxSize
	}
	if conf.MaxSize <= y {
		y = conf.MaxSize
	}

	dst := image.NewRGBA(image.Rect(0, 0, x, y))
	graphics.Thumbnail(dst, img)

	switch mimetype {
	case "jpeg":
		return jpeg.Encode(w, dst, &jpeg.Options{jpeg.DefaultQuality})
	case "png":
		return png.Encode(w, dst)
	default:
		return fmt.Errorf("Crop: mimetype '%s' can't be processed.", mimetype)
	}
}

// *** Resize given image file & write it to io.Writer
func resize(w io.Writer, r io.Reader, size []int) error {
	if size == nil {
		io.Copy(w, r)
		return nil
	}

	img, mimetype, err := image.Decode(r)
	if err != nil {
		io.Copy(w, r)
		return err
	}

	ib := img.Bounds()

	// fit to actual size
	var x, y int = ib.Dx(), ib.Dy()
	if ib.Dx() >= size[0] {
		x = size[0]
	}
	if ib.Dy() >= size[1] {
		y = size[1]
	}

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

	switch mimetype {
	case "jpeg":
		return jpeg.Encode(w, dst, &jpeg.Options{jpeg.DefaultQuality})
	case "png":
		return png.Encode(w, dst)
	default:
		return fmt.Errorf("Resize: mimetype '%s' can't be processed.", mimetype)
	}
}
