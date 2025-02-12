package app

import (
	"bytes"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehlog"
	"go.uber.org/zap"
	"golang.org/x/image/bmp"
	"image/png"
	"io"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"time"
)

type TraceConfig struct {
	TurnPolicy                 string
	TurdSize                   int
	AlphaMax                   float64
	CurveOptimizationTolerance float64
	BlackLevel                 float64
	Invert                     bool
}

var allowedTraceTurnPolicies = []string{"black", "white", "right", "left", "minority", "majority", "random"}

func traceConfigFromQueryParams(c *gin.Context) (TraceConfig, error) {
	config := TraceConfig{
		TurnPolicy:                 "minority",
		TurdSize:                   10_000,
		AlphaMax:                   1,
		CurveOptimizationTolerance: 0.2,
		BlackLevel:                 .5,
		Invert:                     false,
	}

	var err error
	// Parse turn policy.
	if v := c.Query("trace_turn_policy"); v != "" {
		if !slices.Contains(allowedTraceTurnPolicies, v) {
			return TraceConfig{}, meh.NewBadInputErr(fmt.Sprintf("unsupported trace turn policy: %s", v),
				meh.Details{"allowed": allowedTraceTurnPolicies})
		}
		config.TurnPolicy = v
	}

	// Parse turd size.
	if v := c.Query("trace_turd_size"); v != "" {
		config.TurdSize, err = strconv.Atoi(v)
		if err != nil {
			return TraceConfig{}, meh.NewBadInputErrFromErr(err, "parse turd size", meh.Details{"was": v})
		}
		config.TurdSize = min(config.TurdSize, 100_000_000)
		config.TurdSize = max(config.TurdSize, 0)
	}

	// Parse curve optimization tolerance.
	if v := c.Query("trace_alpha_max"); v != "" {
		config.AlphaMax, err = strconv.ParseFloat(v, 64)
		if err != nil {
			return TraceConfig{}, meh.NewBadInputErrFromErr(err, "parse alpha max", meh.Details{"was": v})
		}
		config.AlphaMax = min(config.AlphaMax, 1.5)
		config.AlphaMax = max(config.AlphaMax, 0)
	}

	// Parse curve optimization tolerance.
	if v := c.Query("trace_curve_optimization_tolerance"); v != "" {
		config.CurveOptimizationTolerance, err = strconv.ParseFloat(v, 64)
		if err != nil {
			return TraceConfig{}, meh.NewBadInputErrFromErr(err, "parse curve optimization tolerance", meh.Details{"was": v})
		}
		config.CurveOptimizationTolerance = min(config.CurveOptimizationTolerance, 100_000_000)
		config.CurveOptimizationTolerance = max(config.CurveOptimizationTolerance, 0)
	}

	// Parse black level.
	if v := c.Query("black_level"); v != "" {
		config.BlackLevel, err = strconv.ParseFloat(v, 64)
		if err != nil {
			return TraceConfig{}, meh.NewBadInputErrFromErr(err, "parse black level", meh.Details{"was": v})
		}
		config.BlackLevel = min(config.BlackLevel, 1)
		config.BlackLevel = max(config.BlackLevel, 0)
	}

	// Parse invert.
	if v := c.Query("invert"); v != "" {
		config.Invert, err = strconv.ParseBool(v)
		if err != nil {
			return TraceConfig{}, meh.NewBadInputErrFromErr(err, "parse invert", meh.Details{"was": v})
		}
	}

	return config, nil
}

func (app *App) traceWithPotrace(ctx context.Context, logger *zap.Logger, config TraceConfig, r io.Reader, w io.Writer) error {
	// Parse PNG from reader.
	logger.Debug("read png")
	imgPNG, err := png.Decode(r)
	if err != nil {
		return fmt.Errorf("decode png image: %w", err)
	}

	// Convert to BMP.
	logger.Debug("convert to bmp")
	var imgBMP bytes.Buffer
	err = bmp.Encode(&imgBMP, imgPNG)
	if err != nil {
		return fmt.Errorf("encode image to bmp: %w", err)
	}

	logger.Debug("write to temporary file")
	tmpInputFile, err := os.CreateTemp(os.TempDir(), "tmp-in.*.bmp")
	if err != nil {
		return meh.NewInternalErrFromErr(err, "create temporary input file", nil)
	}
	tmpInputFilename := tmpInputFile.Name()
	defer func() {
		err := os.Remove(tmpInputFilename)
		if err != nil {
			mehlog.Log(logger, meh.NewInternalErrFromErr(err, "remove temporary input file", meh.Details{"tmp_input_filename": tmpInputFilename}))
		}
	}()
	defer func() { _ = tmpInputFile.Close() }()

	// Write BMP to FS.
	n, err := io.Copy(tmpInputFile, &imgBMP)
	if err != nil {
		return meh.NewInternalErrFromErr(err, "write bmp to temporary file", meh.Details{"filename": tmpInputFilename})
	}
	logger.Debug("temporary bmp file written", zap.Int64("bytes", n), zap.String("filename", tmpInputFilename))

	// Create temporary output file.
	logger.Debug("write to temporary file")
	tmpOutputFile, err := os.CreateTemp(os.TempDir(), "tmp-out.*.svg")
	if err != nil {
		return meh.NewInternalErrFromErr(err, "create temporary output file", nil)
	}
	tmpOutputFilename := tmpOutputFile.Name()
	defer func() {
		err := os.Remove(tmpOutputFilename)
		if err != nil {
			mehlog.Log(logger, meh.NewInternalErrFromErr(err, "remove temporary output file", meh.Details{"tmp_output_filename": tmpInputFilename}))
		}
	}()
	_ = tmpOutputFile.Close()

	// Run potrace.
	cmd := exec.CommandContext(ctx, app.config.PotraceFilename)
	cmd.Args = []string{
		"--progress",
		"--output=" + tmpOutputFilename,
		"--backend=svg",
		"--group",
		"--flat",
		fmt.Sprintf("--alphamax=%.10f", config.AlphaMax),
		fmt.Sprintf("--turdsize=%d", config.TurdSize),
		"--turnpolicy=" + config.TurnPolicy,
		fmt.Sprintf("--blacklevel=%.10f", config.BlackLevel),
		"--fill=#ffffff",
	}
	if config.CurveOptimizationTolerance == 0 {
		cmd.Args = append(cmd.Args, "--longcurve")
	} else {
		cmd.Args = append(cmd.Args, fmt.Sprintf("--opttolerance=%.10f", config.CurveOptimizationTolerance))
	}
	if config.Invert {
		cmd.Args = append(cmd.Args, "--invert")
	}
	cmd.Args = append(cmd.Args, tmpInputFilename)
	start := time.Now()
	logger.Debug("run tracing",
		zap.String("command", app.config.PotraceFilename),
		zap.Strings("args", cmd.Args),
		zap.Time("start_at", start))
	got, err := cmd.CombinedOutput()
	logger.Debug("output", zap.ByteString("output", got))
	if err != nil {
		return meh.NewInternalErrFromErr(err, "run potrace", meh.Details{
			"potrace_filename": app.config.PotraceFilename,
			"args":             cmd.Args,
		})
	}
	logger.Debug("potrace done", zap.Duration("took", time.Since(start)))

	// Read file and write to writer.
	tmpOutputFile, err = os.Open(tmpOutputFilename)
	if err != nil {
		return meh.NewInternalErrFromErr(err, "open temporary output file", meh.Details{"filename": tmpOutputFilename})
	}
	defer func() { _ = tmpOutputFile.Close() }()
	n, err = io.Copy(w, tmpOutputFile)
	if err != nil {
		return meh.NewInternalErrFromErr(err, "copy tmp output file contents to writer", meh.Details{"filename": tmpOutputFilename})
	}
	logger.Debug("finished reading output file", zap.Int64("bytes", n))

	return nil
}
