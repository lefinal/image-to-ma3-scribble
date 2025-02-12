package app

import (
	"encoding/xml"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lefinal/image-to-ma3-scribble/scribble"
	"github.com/lefinal/meh"
	"go.uber.org/zap"
	"image/color"
	"io"
	"strconv"
	"strings"
)

const MaxSVGPathSegments = 100

type MA3ScribbleConfig struct {
	Name string
	// StrokeThickness from 0.0 to 10.0.
	StrokeThickness float64
	StrokeColor     color.RGBA
}

func ma3ScribbleConfigFromQueryParams(c *gin.Context) (MA3ScribbleConfig, error) {
	config := MA3ScribbleConfig{
		Name:            "MyScribble",
		StrokeThickness: .2,
		StrokeColor:     color.RGBA{R: 255, G: 255, B: 255, A: 255},
	}

	var err error

	// Parse name.
	if v := c.Query("ma3_scribble_name"); v != "" {
		config.Name = v
		if len(config.Name) > 1000 {
			return MA3ScribbleConfig{}, meh.NewBadInputErr("name exceeded max length", meh.Details{"was": v})
		}
	}

	// Parse stroke thickness.
	if v := c.Query("ma3_scribble_stroke_thickness"); v != "" {
		config.StrokeThickness, err = strconv.ParseFloat(v, 64)
		if err != nil {
			return MA3ScribbleConfig{}, meh.NewBadInputErrFromErr(err, "parse stroke thickness", meh.Details{"was": v})
		}
		config.StrokeThickness = min(config.StrokeThickness, 10.0)
		config.StrokeThickness = max(config.StrokeThickness, 0)
	}

	// Parse stroke color.
	if v := c.Query("ma3_scribble_stroke_color"); v != "" {
		config.StrokeColor, err = parseHexRGBA(v)
		if err != nil {
			return MA3ScribbleConfig{}, meh.NewBadInputErrFromErr(err, "parse stroke color", meh.Details{"was": v})
		}
	}

	return config, nil
}

func parseHexRGBA(hex string) (color.RGBA, error) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 8 {
		return color.RGBA{}, meh.NewBadInputErr("invalid hex length: must be 8 characters", nil)
	}

	r, err := strconv.ParseUint(hex[0:2], 16, 8)
	if err != nil {
		return color.RGBA{}, meh.NewBadInputErrFromErr(err, "invalid red value", nil)
	}

	g, err := strconv.ParseUint(hex[2:4], 16, 8)
	if err != nil {
		return color.RGBA{}, meh.NewBadInputErrFromErr(err, "invalid green value", nil)
	}

	b, err := strconv.ParseUint(hex[4:6], 16, 8)
	if err != nil {
		return color.RGBA{}, meh.NewBadInputErrFromErr(err, "invalid blue value", nil)
	}

	a, err := strconv.ParseUint(hex[6:8], 16, 8)
	if err != nil {
		return color.RGBA{}, meh.NewBadInputErrFromErr(err, "invalid alpha value", nil)
	}

	return color.RGBA{
		R: uint8(r),
		G: uint8(g),
		B: uint8(b),
		A: uint8(a),
	}, nil
}

func rgbaToHex(c color.RGBA) string {
	r := int(c.R)
	g := int(c.G)
	b := int(c.B)
	a := int(c.A)

	// Convert to hex and pad with 0 if necessary.  %02x formats with two hex digits, padding with zero.
	hexColor := fmt.Sprintf("#%02x%02x%02x%02x", r, g, b, a)

	return strings.ToUpper(hexColor) // Or strings.ToLower if you prefer lowercase
}

// Define structures to capture the SVG and path data
type SVG struct {
	XMLName xml.Name   `xml:"svg"`
	Width   string     `xml:"width,attr"`
	Height  string     `xml:"height,attr"`
	Groups  []SVGGroup `xml:"g"`
}

type SVGGroup struct {
	XMLName   xml.Name  `xml:"g"`
	Transform string    `xml:"transform,attr"`
	Paths     []SVGPath `xml:"path"`
}

type SVGPath struct {
	D string `xml:"d,attr"` // The path data (d attribute)
}

type ma3ScribblePathBuilder interface {
	feed(normalizedRelative float64) bool
	build() []float64
	nextXY() (float64, float64)
	newEmpty(startX, startY float64) ma3ScribblePathBuilder
}

type ma3ScribblePathLineBuilder struct {
	n int

	startX    float64
	startY    float64
	controlX1 float64
	controlY1 float64
	controlX2 float64
	controlY2 float64
	endX      float64
	endY      float64
}

func (builder *ma3ScribblePathLineBuilder) newEmpty(startX, startY float64) ma3ScribblePathBuilder {
	return &ma3ScribblePathLineBuilder{
		startX: startX,
		startY: startY,
	}
}

func (builder *ma3ScribblePathLineBuilder) feed(normalizedRelative float64) bool {
	switch builder.n {
	case 0:
		builder.endX = builder.startX + normalizedRelative
	case 1:
		builder.endY = builder.startY - normalizedRelative
		builder.controlX1 = (builder.startX + builder.endX) / 2
		builder.controlY1 = (builder.startY + builder.endY) / 2
		builder.controlX2 = (builder.startX + builder.endX) / 2
		builder.controlY2 = (builder.startY + builder.endY) / 2
		return true
	}
	builder.n++
	return false
}

func (builder *ma3ScribblePathLineBuilder) build() []float64 {
	return []float64{
		builder.startX,
		builder.startY,
		builder.controlX1,
		builder.controlY1,
		builder.controlX2,
		builder.controlY2,
		builder.endX,
		builder.endY,
	}
}

func (builder *ma3ScribblePathLineBuilder) nextXY() (float64, float64) {
	return builder.endX, builder.endY
}

type ma3ScribblePathCubicBuilder struct {
	n int

	startX    float64
	startY    float64
	controlX1 float64
	controlY1 float64
	controlX2 float64
	controlY2 float64
	endX      float64
	endY      float64
}

func (builder *ma3ScribblePathCubicBuilder) newEmpty(startX, startY float64) ma3ScribblePathBuilder {
	return &ma3ScribblePathCubicBuilder{
		startX: startX,
		startY: startY,
	}
}

func (builder *ma3ScribblePathCubicBuilder) feed(normalizedRelative float64) bool {
	switch builder.n {
	case 0:
		builder.controlX1 = builder.startX + normalizedRelative
	case 1:
		builder.controlY1 = builder.startY + normalizedRelative
	case 2:
		builder.controlX2 = builder.startX + normalizedRelative
	case 3:
		builder.controlY2 = builder.startY + normalizedRelative
	case 4:
		builder.endX = builder.startX + normalizedRelative
	case 5:
		builder.endY = builder.startY + normalizedRelative
		return true
	}
	builder.n++
	return false
}

func (builder *ma3ScribblePathCubicBuilder) build() []float64 {
	return []float64{
		builder.startX,
		builder.startY,
		builder.controlX1,
		builder.controlY1,
		builder.controlX2,
		builder.controlY2,
		builder.endX,
		builder.endY,
	}
}

func (builder *ma3ScribblePathCubicBuilder) nextXY() (float64, float64) {
	return builder.endX, builder.endY
}

type transformOptions struct {
	translateX float64
	translateY float64
	scaleX     float64
	scaleY     float64
}

func transformOptionsFromString(s string) (transformOptions, error) {
	opts := transformOptions{
		translateX: 0,
		translateY: 0,
		scaleX:     0,
		scaleY:     0,
	}
	actions := strings.Split(s, " ")
	for _, action := range actions {
		actionType := strings.Split(action, "(")[0]
		param1Str := strings.Split(strings.Split(action, "(")[1], ",")[0]
		param2Str := strings.Split(strings.Split(strings.Split(action, "(")[1], ",")[1], ")")[0]

		param1, err := strconv.ParseFloat(param1Str, 64)
		if err != nil {
			return transformOptions{}, meh.NewInternalErrFromErr(err, "parse param 1 from action", meh.Details{"action": action})
		}
		param2, err := strconv.ParseFloat(param2Str, 64)
		if err != nil {
			return transformOptions{}, meh.NewInternalErrFromErr(err, "parse param 2 from action", meh.Details{"action": action})
		}

		switch actionType {
		case "translate":
			opts.translateX = param1
			opts.translateY = param2
		case "scale":
			opts.scaleX = param1
			opts.scaleY = param2
		}
	}
	return opts, nil
}

func (app *App) encodeSVGToMA3Scribble(logger *zap.Logger, config MA3ScribbleConfig, svgRaw io.Reader, w io.Writer) error {
	// Parse the SVG file
	var svg SVG
	decoder := xml.NewDecoder(svgRaw)
	err := decoder.Decode(&svg)
	if err != nil {
		return meh.NewInternalErrFromErr(err, "parse svg", nil)
	}

	viewBoxHeight, err := strconv.ParseFloat(strings.TrimSuffix(svg.Height, "pt"), 64)
	if err != nil {
		return meh.NewInternalErrFromErr(err, "parse svg height", meh.Details{"was": svg.Height})
	}
	viewBoxWidth, err := strconv.ParseFloat(strings.TrimSuffix(svg.Width, "pt"), 64)
	if err != nil {
		return meh.NewInternalErrFromErr(err, "parse svg width", meh.Details{"was": svg.Width})
	}

	// Calculate thickness in MA3 scribble format.
	ma3Thickness := strokeThicknessToScribbleFormat(config.StrokeThickness)

	// Output the paths
	ma3Paths := make([]string, 0)

	for _, group := range svg.Groups {
		transformOptions, err := transformOptionsFromString(group.Transform)
		if err != nil {
			return meh.Wrap(err, "parse transform options", meh.Details{"was": group.Transform})
		}
		logger.Debug("building from paths", zap.Int("path_count", len(group.Paths)))
		for _, path := range group.Paths {

			scribbleLineSegments := make([]string, 0)
			// Write color.
			scribbleLineSegments = append(scribbleLineSegments, rgbaToHex(config.StrokeColor)[1:])
			// Write thickness.
			scribbleLineSegments = append(scribbleLineSegments, fmt.Sprintf("%.6f", config.StrokeThickness*0.01))
			// Write segments.
			path.D = strings.TrimSpace(path.D)
			path.D = strings.ReplaceAll(path.D, "\n", " ")
			path.D = strings.ReplaceAll(path.D, "\r", "")
			path.D = strings.ReplaceAll(path.D, "\t", "")
			path.D = strings.ReplaceAll(path.D, "M", "")
			path.D = strings.ReplaceAll(path.D, "z", "")
			path.D = strings.ReplaceAll(path.D, "c", "c ")
			path.D = strings.ReplaceAll(path.D, "l", "l ")
			pathSegments := strings.Split(path.D, " ")
			// Ignore first (is 'M').

			var currentBuilder ma3ScribblePathBuilder
			var currentX, currentY float64
			for segmentIdx, segment := range pathSegments {
				// Check if line- or curve-indicator. This will not happen when we haven't read
				// the first move-command yet.
				if segment == "c" {
					currentBuilder = &ma3ScribblePathCubicBuilder{
						startX: currentX,
						startY: currentY,
					}
					continue
				} else if segment == "l" {
					currentBuilder = &ma3ScribblePathLineBuilder{
						startX: currentX,
						startY: currentY,
					}
					continue
				}

				// Parse the number.
				n, err := strconv.Atoi(segment)
				if err != nil {
					return meh.NewInternalErrFromErr(err, "parse number segment", meh.Details{"was": segment})
				}
				normalized := float64(n)

				// Read initial move-command.
				switch {
				case segmentIdx == 0:
					currentX = normalized
					continue
				case segmentIdx == 1:
					currentY = normalized
					continue
				}

				// Feed number.
				if currentBuilder == nil {
					return meh.NewInternalErr("no builder", meh.Details{"segment_idx": segmentIdx})
				}
				if currentBuilder.feed(normalized) {
					// Builder ready.
					result := currentBuilder.build()
					resultAsStrings := []string{
						rgbaToHex(config.StrokeColor)[1:],
						fmt.Sprintf("%.6f", ma3Thickness),
					}

					// Determine scaling factor.
					largerDimension := max(viewBoxWidth, viewBoxHeight)
					scaleFactor := 1.0 / largerDimension

					normalizedWidth := viewBoxWidth * scaleFactor
					normalizedHeight := viewBoxHeight * scaleFactor

					// Calculate offset for centering.
					var xOffset, yOffset float64
					if viewBoxWidth > viewBoxHeight {
						xOffset = 0
						yOffset = (1 - normalizedHeight) / 2.0
					} else {
						xOffset = (1 - normalizedWidth) / 2.0
						yOffset = 0
					}

					for i, f := range result {
						// Apply scaling and fitting.
						if i%2 == 1 {
							// Y coordinate. Shift btw.
							f *= transformOptions.scaleY
							f = f*scaleFactor + yOffset
							f += normalizedHeight
						} else {
							f *= transformOptions.scaleX
							f = f*scaleFactor + xOffset
						}

						resultAsStrings = append(resultAsStrings, fmt.Sprintf("%.6f", f))
					}
					ma3Paths = append(ma3Paths, strings.Join(resultAsStrings, ","))
					currentX, currentY = currentBuilder.nextXY()
					currentBuilder = currentBuilder.newEmpty(currentX, currentY)
				}
			}
		}
	}

	// Build the actual MA3 scribble.
	ma3Scribble := scribble.New(config.Name, ma3Paths)
	enc := xml.NewEncoder(w)
	return enc.Encode(ma3Scribble)
}

func strokeThicknessToScribbleFormat(thickness float64) float64 {
	return thickness/10.0*(scribble.ScribbleMaxThickness-scribble.ScribbleMinThickness) + scribble.ScribbleMinThickness
}
