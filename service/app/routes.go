package app

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lefinal/image-to-ma3-scribble/web"
	"github.com/lefinal/meh"
	"go.uber.org/zap"
	"net/http"
	"os"
	"strings"
)

func (app *App) handlePNGToMA3Scribble(previewOnly bool) web.HandlerFunc {
	return func(logger *zap.Logger, c *gin.Context) error {
		// Parse query params.
		preprocessOptions, err := preprocessPNGOptionsFromQueryParams(c)
		if err != nil {
			return meh.Wrap(err, "parse preprocess options from query params", nil)
		}
		traceConfig, err := traceConfigFromQueryParams(c)
		if err != nil {
			return meh.Wrap(err, "parse trace request config from query params", nil)
		}
		ma3ScribbleConfig, err := ma3ScribbleConfigFromQueryParams(c)
		if err != nil {
			return meh.Wrap(err, "parse ma3 scribble config from query params", nil)
		}

		// Preprocess.
		var preprocessedPNG bytes.Buffer
		err = app.preprocessPNG(logger.Named("preprocess"), c.Request.Body, &preprocessedPNG, preprocessOptions)
		if err != nil {
			return meh.Wrap(err, "preprocess png", nil)
		}

		// Trace.
		var tracedSVG bytes.Buffer
		err = app.traceWithPotrace(c.Request.Context(), logger.Named("trace"), traceConfig, &preprocessedPNG, &tracedSVG)
		if err != nil {
			return meh.Wrap(err, "trace svg with potrace", nil)
		}

		if previewOnly {
			// Make some sneaky changes to simulate stroke settings.
			tracedSVGStr := tracedSVG.String()
			svgStrokeColor := rgbaToHex(ma3ScribbleConfig.StrokeColor)
			replacement := fmt.Sprintf(`fill="transparent" stroke="%s" stroke-width="50"`, svgStrokeColor)
			tracedSVGStr = strings.ReplaceAll(tracedSVGStr, `fill="#000000" stroke="none"`, replacement)
			tracedSVGStr = strings.ReplaceAll(tracedSVGStr, `fill="#ffffff" stroke="none"`, replacement)

			// We just keep it lol.
			_ = os.WriteFile("TMPx.svg", tracedSVG.Bytes(), 0644)

			c.Data(http.StatusOK, "image/xml+svg", []byte(tracedSVGStr))
			return nil
		}

		// Encode to MA3 scribble.
		var ma3ScribbleXML bytes.Buffer
		err = app.encodeSVGToMA3Scribble(logger.Named("encode-ma3"), ma3ScribbleConfig, &tracedSVG, &ma3ScribbleXML)
		if err != nil {
			return meh.Wrap(err, "encode svg to ma3 scribble", nil)
		}

		c.Data(http.StatusOK, "application/xml", ma3ScribbleXML.Bytes())
		return nil
	}
}
