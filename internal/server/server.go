package server

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/brpaz/echozap"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strings"
	"time"
)

var ErrSignatureVerificationFailed = errors.New("event signature verification failed")

type Callback func(ctx echo.Context, presentedEventType string, body []byte, logger *zap.SugaredLogger) error

type Server struct {
	eventTypesSet mapset.Set[string]
	callback      Callback
	logger        *zap.SugaredLogger
}

func New(callback Callback, logger *zap.SugaredLogger) *Server {
	return &Server{
		eventTypesSet: mapset.NewSet[string](eventTypes...),
		callback:      callback,
		logger:        logger,
	}
}

func (server *Server) Run(ctx context.Context) error {
	// Configure HTTP server
	e := echo.New()

	e.Use(echozap.ZapLogger(server.logger.Desugar()))

	e.POST(httpPath, server.handler)

	httpServer := &http.Server{
		Addr:              httpAddr,
		Handler:           e,
		ReadHeaderTimeout: 10 * time.Second,
	}

	server.logger.Infof("starting HTTP server on %s", httpAddr)

	httpServerErrCh := make(chan error, 1)

	go func() {
		httpServerErrCh <- httpServer.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		if err := httpServer.Close(); err != nil {
			return err
		}

		return ctx.Err()
	case httpServerErr := <-httpServerErrCh:
		return httpServerErr
	}
}

func (server *Server) handler(ctx echo.Context) error {
	// Make sure that this is an event we've been looking for
	presentedEventType := ctx.Request().Header.Get("X-Cirrus-Event")

	if server.eventTypesSet.Cardinality() != 0 && !server.eventTypesSet.Contains(presentedEventType) {
		server.logger.Debugf("skipping event of type %q because we only process events of types %s",
			presentedEventType, strings.Join(server.eventTypesSet.ToSlice(), ", "))

		return ctx.NoContent(http.StatusOK)
	}

	// Verify that this event comes from the Cirrus CI
	body, err := io.ReadAll(ctx.Request().Body)
	if err != nil {
		server.logger.Warnf("failed to read request's body: %v", err)

		return ctx.NoContent(http.StatusBadRequest)
	}

	if err := verifyEvent(ctx, body); err != nil {
		server.logger.Warnf("%v", err)

		return ctx.NoContent(http.StatusBadRequest)
	}

	if err := server.callback(ctx, presentedEventType, body, server.logger); err != nil {
		server.logger.Warnf("%v", err)

		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.NoContent(http.StatusCreated)
}

func verifyEvent(ctx echo.Context, body []byte) error {
	// Nothing to do
	if secretToken == "" {
		return nil
	}

	// Calculate the expected signature
	hmacSHA256 := hmac.New(sha256.New, []byte(secretToken))
	hmacSHA256.Write(body)
	expectedSignature := hmacSHA256.Sum(nil)

	// Prepare the presented signature
	presentedSignatureRaw := ctx.Request().Header.Get("X-Cirrus-Signature")
	presentedSignature, err := hex.DecodeString(presentedSignatureRaw)
	if err != nil {
		return fmt.Errorf("%w: failed to hex-decode the signature %q: %v",
			ErrSignatureVerificationFailed, presentedSignatureRaw, err)
	}

	// Compare signatures
	if !hmac.Equal(expectedSignature, presentedSignature) {
		return fmt.Errorf("%w: signature is not valid", ErrSignatureVerificationFailed)
	}

	return ctx.NoContent(http.StatusCreated)
}
