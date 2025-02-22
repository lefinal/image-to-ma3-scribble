package app

import (
	"github.com/disintegration/gift"
	"github.com/gin-gonic/gin"
	"github.com/lefinal/meh"
	"go.uber.org/zap"
	"image"
	"image/color"
	"image/png"
	"io"
	"math"
	"strconv"
	"time"
)

type preprocessPNGOptions struct {
	TransparencyReplacementColor color.RGBA
	BlurRadius                   float32
}

func preprocessPNGOptionsFromQueryParams(c *gin.Context) (preprocessPNGOptions, error) {
	options := preprocessPNGOptions{
		TransparencyReplacementColor: color.RGBA{R: 255, G: 255, B: 255, A: 255},
		BlurRadius:                   0.0,
	}

	var err error
	// Parse replacement color.
	if v := c.Query("preprocess_transparency_replacement_color"); v != "" {
		options.TransparencyReplacementColor, err = parseHexRGBA(v)
		if err != nil {
			return preprocessPNGOptions{}, meh.NewBadInputErrFromErr(err, "parse transparency replacement color", meh.Details{"was": v})
		}
		options.TransparencyReplacementColor.A = 255
	}

	// Parse blur radius.
	if v := c.Query("preprocess_blur_radius"); v != "" {
		f, err := strconv.ParseFloat(v, 32)
		if err != nil {
			return preprocessPNGOptions{}, meh.NewBadInputErrFromErr(err, "parse blur radius", meh.Details{"was": v})
		}
		options.BlurRadius = float32(f)
	}

	return options, nil
}

func (app *App) preprocessPNG(logger *zap.Logger, r io.Reader, w io.Writer, options preprocessPNGOptions) error {
	start := time.Now()
	logger.Debug("start preprocessing")
	defer func() {
		logger.Debug("finished preprocessing", zap.Duration("took", time.Since(start)))
	}()

	pngImage, err := png.Decode(r)
	if err != nil {
		return meh.NewBadInputErrFromErr(err, "parse png", nil)
	}

	// Recolor transparency.
	bounds := pngImage.Bounds()
	newImage := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			original := pngImage.At(x, y)
			r, g, b, a := original.RGBA() // Returns values in range 0-65535

			if a < math.MaxUint16 { // If the pixel is not fully opaque
				newImage.Set(x, y, options.TransparencyReplacementColor)
			} else {
				newImage.Set(x, y, color.RGBA{
					R: uint8(r >> 8),
					G: uint8(g >> 8),
					B: uint8(b >> 8),
					A: 255})
			}
		}
	}

	// Apply Gaussian Blur.
	if options.BlurRadius > 0 {
		blurredImage := image.NewRGBA(bounds)
		g := gift.New(gift.GaussianBlur(options.BlurRadius))
		g.Draw(blurredImage, newImage)
		newImage = blurredImage
	}

	// Encode PNG.
	err = png.Encode(w, newImage)
	if err != nil {
		return meh.NewInternalErrFromErr(err, "encode png", nil)
	}
	return nil
}
