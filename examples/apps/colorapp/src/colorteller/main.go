package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/kyawmyintthein/aws-app-mesh-examples/colorapp/teller/rpc/service"
	"github.com/sirupsen/logrus"
	"github.com/twitchtv/twirp"
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

func NewColorHandler() service.ColortellerService {
	return &colorHandler{}
}

func (h *colorHandler) GetColor(ctx context.Context, empty *service.Empty) (*service.Message, error) {
	return &service.Message{Value: getColor()}, nil
}

func (h *colorHandler) GetStage(ctx context.Context, empty *service.Empty) (*service.Message, error) {
	return &service.Message{Value: getStage()}, nil
}

func (h *colorHandler) Ping(ctx context.Context, empty *service.Empty) (*service.Empty, error) {
	log.Println("ping requested, reponding with HTTP 200")
	return &service.Empty{}, nil
}

func main() {
	log.Println("starting server, listening on port " + getServerPort())
	server := NewColorHandler()
	twirpHandler := service.NewColortellerServiceServer(server, NewXrayServerHooks())
	xraySegmentNamer := xray.NewFixedSegmentNamer(fmt.Sprintf("%s-colorteller-%s", getStage(), getColor()))
	http.ListenAndServe(":"+getServerPort(), xray.Handler(xraySegmentNamer, twirpHandler))
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
