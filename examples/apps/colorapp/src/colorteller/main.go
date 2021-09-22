package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/kyawmyintthein/aws-app-mesh-examples/colorapp/teller/rpc/colorteller"
	"github.com/sirupsen/logrus"
	"github.com/twitchtv/twirp"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

const defaultPort = "8080"
const defaultColor = "black"
const defaultStage = "default"

func getServerPort() string {
	port := os.Getenv("SERVER_PORT")
	if port != "" {
		return port
	}

	return defaultPort
}

func getColor() string {
	color := os.Getenv("COLOR")
	if color != "" {
		return color
	}

	return defaultColor
}

func getStage() string {
	stage := os.Getenv("STAGE")
	if stage != "" {
		return stage
	}

	return defaultStage
}

type colorHandler struct{}

func NewColorHandler() colorteller.ColortellerService {
	return &colorHandler{}
}

func (h *colorHandler) GetColor(ctx context.Context, empty *colorteller.Empty) (*colorteller.Message, error) {
	return &colorteller.Message{Value: getColor()}, nil
}

func (h *colorHandler) GetStage(ctx context.Context, empty *colorteller.Empty) (*colorteller.Message, error) {
	return &colorteller.Message{Value: getStage()}, nil
}

func (h *colorHandler) Ping(ctx context.Context, empty *colorteller.Empty) (*colorteller.Empty, error) {
	log.Println("ping requested, reponding with HTTP 200")
	return &colorteller.Empty{}, nil
}

func main() {
	log.Println("starting server, listening on port " + getServerPort())
	server := NewColorHandler()
	twirpHandler := colorteller.NewColortellerServiceServer(server, twirp.WithServerPathPrefix("/rz"), NewXrayServerHooks())
	xraySegmentNamer := xray.NewFixedSegmentNamer(fmt.Sprintf("%s-colorteller-%s", getStage(), getColor()))
	var cw ConnectionWatcher
	h2s := &http2.Server{}
	h1s := &http.Server{
		Addr:      ":" + getServerPort(),
		Handler:   h2c.NewHandler(xray.Handler(xraySegmentNamer, twirpHandler), h2s),
		ConnState: cw.OnStateChange,
	}
	log.Fatal(h1s.ListenAndServe())
}

func NewXrayServerHooks() *twirp.ServerHooks {

	hooks := &twirp.ServerHooks{}

	hooks.RequestRouted = func(ctx context.Context) (context.Context, error) {

		method, ok := twirp.MethodName(ctx)
		if !ok {
			return ctx, nil
		}

		logrus.WithField("method", method).Info("BeginSegment")

		if xray.GetSegment(ctx).InProgress {
			ctx, _ = xray.BeginSubsegment(ctx, method)
		}

		return ctx, nil
	}

	hooks.ResponseSent = func(ctx context.Context) {
		method, ok := twirp.MethodName(ctx)
		if !ok {
			return
		}
		logrus.WithField("method", method).Info("Close")
		seg := xray.GetSegment(ctx)
		if seg != nil {
			seg.Close(nil)
		}
	}

	return hooks
}

type ConnectionWatcher struct {
	n int64
}

// OnStateChange records open connections in response to connection
// state changes. Set net/http Server.ConnState to this method
// as value.
func (cw *ConnectionWatcher) OnStateChange(conn net.Conn, state http.ConnState) {
	switch state {
	case http.StateNew:
		atomic.AddInt64(&cw.n, 1)
	case http.StateHijacked, http.StateClosed:
		atomic.AddInt64(&cw.n, -1)
	}
	log.Printf("connection count : %d \n", int(atomic.LoadInt64(&cw.n)))
}

// Count returns the number of connections at the time
// the call.
func (cw *ConnectionWatcher) Count() int {
	return int(atomic.LoadInt64(&cw.n))

}
