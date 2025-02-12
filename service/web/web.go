package web

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lefinal/image-to-ma3-scribble/validate"
	"github.com/lefinal/meh"
	"github.com/lefinal/meh/mehhttp"
	"github.com/lefinal/meh/mehlog"
	"go.uber.org/zap"
	"net/http"
	"time"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
	mehhttp.SetHTTPStatusCodeMapping(func(code meh.Code) int {
		//nolint:exhaustive
		switch code {
		case meh.ErrNotFound:
			return http.StatusNotFound
		case meh.ErrBadInput:
			return http.StatusBadRequest
		case meh.ErrForbidden:
			return http.StatusForbidden
		case meh.ErrUnauthorized:
			return http.StatusUnauthorized
		default:
			return http.StatusInternalServerError
		}
	})
}

// HandlerFunc describes the signature of a function that handles HTTP requests
// with HandlerBuilder.GinHandler.
type HandlerFunc func(logger *zap.Logger, c *gin.Context) error

// HandlerBuilder provides GinHandler for creating a gin.HandlerFunc.
type HandlerBuilder struct {
	Logger *zap.Logger
}

type errorResponse struct {
	Title         string                      `json:"title"`
	Status        int                         `json:"status"`
	Details       string                      `json:"detail"`
	InvalidFields []errorResponseInvalidField `json:"invalidFields,omitempty"`
}

type errorResponseInvalidField struct {
	Field          string   `json:"field"`
	Message        string   `json:"message"`
	Code           string   `json:"code"`
	ValidationCode string   `json:"validationCode"`
	Arguments      []string `json:"arguments"`
}

// GinHandler is a method that returns a gin.HandlerFunc for handling requests
// with a given HandlerFunc. It calls the provided HandlerFunc and handles any
// error that may occur. If an error occurs, it logs the error and sends a proper
// error response back to the client. The error response includes the error
// details as well as the status code specified in the error.
func (b *HandlerBuilder) GinHandler(fn HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestLogger := b.Logger.Named("request").With(zap.String("http_req_url", c.Request.URL.String()),
			zap.String("http_req_host", c.Request.Host),
			zap.String("http_req_method", c.Request.Method),
			zap.String("http_req_user_agent", c.Request.UserAgent()),
			zap.String("http_req_remote_addr", c.Request.RemoteAddr),
		)
		err := fn(requestLogger, c)
		if err == nil {
			return
		}
		c.Abort()
		e := err
		// Properly respond with error details.
		e = meh.ApplyDetails(e, meh.Details{
			"http_req_url":         c.Request.URL.String(),
			"http_req_host":        c.Request.Host,
			"http_req_method":      c.Request.Method,
			"http_req_user_agent":  c.Request.UserAgent(),
			"http_req_remote_addr": c.Request.RemoteAddr,
		})
		mehlog.Log(requestLogger, e)
		httpStatus := mehhttp.HTTPStatusCode(e)
		response := errorResponse{
			Title:  string(meh.ErrorCode(e)),
			Status: httpStatus,
		}
		if meh.ErrorCode(err) == meh.ErrBadInput {
			response.Details = err.Error()
		}
		responseJSON, err := json.Marshal(response)
		if err != nil {
			mehlog.Log(requestLogger, meh.Wrap(err, "marshal error response body", meh.Details{"was": fmt.Sprintf("%+v", response)}))
			return
		}
		err = respondHTTP(c.Writer, "application/json", responseJSON, httpStatus)
		if err != nil {
			mehlog.Log(requestLogger, meh.Wrap(err, "respond http", meh.Details{
				"status": httpStatus,
			}))
			return
		}
	}
}

// invalidFieldsForError creates an errorResponseInvalidField-list for the given
// error. If the error is nil, nil is also returned. Otherwise, we try to find
// the first error with meh.ErrBadInput and check whether the wrapped error is of
// type validate.IssueList. If so, we construct the appropriate web
// representations from it.
func invalidFieldsForError(e error) []errorResponseInvalidField {
	if e == nil || meh.ErrorCode(e) != meh.ErrBadInput {
		return nil
	}
	foundBadInput := false
	for e != nil {
		mehErr := meh.Cast(e)
		// Check whether we were checking the first actually found bad-input error and
		// now arrived at a wrapped other error. In this case, we ignore it, as it comes
		// from a more nested error, and we don't want to output these fields.
		if foundBadInput && mehErr.Code != meh.ErrNeutral && mehErr.Code != meh.ErrUnexpected {
			return nil
		} else if !foundBadInput {
			foundBadInput = mehErr.Code == meh.ErrBadInput
		}
		// While we are "within" the bad-input error, we try to find a wrapped
		// issue-list.
		if foundBadInput {
			// We found a bad-input error. Check whether we have a wrapped issue-list. That
			// is why we explicitly perform a type-assertion here and don't use errors.As.
			wrappedIssueList, ok := e.(validate.IssueList) //nolint:errorlint
			if ok {
				// Found it.
				invalidFields := make([]errorResponseInvalidField, 0)
				for _, issue := range wrappedIssueList.Issues {
					invalidFields = append(invalidFields, errorResponseInvalidField{
						Field:          issue.Field,
						Message:        issue.Detail,
						Code:           "invalid",
						ValidationCode: "",
						Arguments:      nil,
					})
				}
				return invalidFields
			}
		}
		if mehErr.Code != meh.ErrUnexpected {
			e = mehErr.WrappedErr
		} else {
			e = nil
		}
	}
	// We did not find a nested issue list.
	return nil
}

// respondHTTP responds the given message with the status to the
// http.ResponseWriter.
func respondHTTP(w http.ResponseWriter, contentType string, body []byte, status int) error {
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(status)
	_, err := w.Write(body)
	if err != nil {
		return meh.NewErrFromErr(err, mehhttp.ErrCommunication, "write", nil)
	}
	return nil
}

// RequestDebugLogger logs requests on zap.DebugLevel to the given zap.Logger.
// The idea is based on gin.Logger.
func RequestDebugLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		// Process request.
		c.Next()
		// Log results.
		logger.Debug("request",
			zap.Time("timestamp", start),
			zap.Duration("took", time.Since(start)),
			zap.String("url", c.Request.URL.String()),
			zap.String("client_ip", c.ClientIP()),
			zap.String("method", c.Request.Method),
			zap.Int("status_code", c.Writer.Status()),
			zap.String("error_message", c.Errors.ByType(gin.ErrorTypePrivate).String()),
			zap.Int("body_size", c.Writer.Size()),
			zap.String("user_agent", c.Request.UserAgent()))
	}
}
