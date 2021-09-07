package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/kyawmyintthein/aws-app-mesh-examples/colorapp/gateway/rpc/frontservice"
	"github.com/kyawmyintthein/aws-app-mesh-examples/colorapp/gateway/rpc/service"
	"github.com/pkg/errors"
	"golang.org/x/net/http2"

	"github.com/sirupsen/logrus"
	"github.com/twitchtv/twirp"
)

const defaultPort = "8080"
const defaultStage = "default"
const maxColors = 1000

var colors [maxColors]string
var colorsIdx int
var colorsMutext = &sync.Mutex{}

func getServerPort() string {
	port := os.Getenv("SERVER_PORT")
	if port != "" {
		return port
	}

	return defaultPort
}

func getStage() string {
	stage := os.Getenv("STAGE")
	if stage != "" {
		return stage
	}

	return defaultStage
}

func getColorTellerEndpoint() (string, error) {
	colorTellerEndpoint := os.Getenv("COLOR_TELLER_ENDPOINT")
	if colorTellerEndpoint == "" {
		return "", errors.New("COLOR_TELLER_ENDPOINT is not set")
	}
	return colorTellerEndpoint, nil
}

type colorHandler struct{}

func NewColorHandler() frontservice.GatewayService {
	return &colorHandler{}
}

func (h *colorHandler) GetColor(ctx context.Context, empty *frontservice.Empty) (*frontservice.Message, error) {
	color, err := getColorFromColorTeller(ctx)
	if err != nil {
		return &frontservice.Message{Color: color, Error: err.Error()}, err
	}

	colorsMutext.Lock()
	defer colorsMutext.Unlock()

	addColor(color)
	statsJson, err := json.Marshal(getRatios())
	if err != nil {
		return &frontservice.Message{Color: color, Error: err.Error()}, err
	}
	return &frontservice.Message{Color: color, Stats: string(statsJson)}, err
}

func addColor(color string) {
	colors[colorsIdx] = color

	colorsIdx += 1
	if colorsIdx >= maxColors {
		colorsIdx = 0
	}
}

func getRatios() map[string]float64 {
	counts := make(map[string]int)
	var total = 0

	for _, c := range colors {
		if c != "" {
			counts[c] += 1
			total += 1
		}
	}

	ratios := make(map[string]float64)
	for k, v := range counts {
		ratio := float64(v) / float64(total)
		ratios[k] = math.Round(ratio*100) / 100
	}

	return ratios
}

func getColorFromColorTeller(ctx context.Context) (string, error) {
	colorTellerEndpoint, err := getColorTellerEndpoint()
	if err != nil {
		return "-n/a-", err
	}
	httpclient := &http.Client{
		Transport: &http2.Transport{
			// Allow non-https urls
			AllowHTTP: true,
			// Make the transport *not-actually* use TLS
			DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
				return net.Dial(network, addr)
			},
		},
	}
	client := xray.Client(httpclient)
	colorTellerClient := service.NewColortellerServiceProtobufClient(fmt.Sprintf("http://%s", colorTellerEndpoint), client)
	msg, err := colorTellerClient.GetColor(ctx, &service.Empty{})
	if err != nil {
		return "-n/a-", err
	}

	color := strings.TrimSpace(string(msg.Value))
	if len(color) < 1 {
		return "-n/a-", errors.New("Empty response from colorTeller")
	}

	return color, nil
}

func getTCPEchoEndpoint() (string, error) {
	tcpEchoEndpoint := os.Getenv("TCP_ECHO_ENDPOINT")
	if tcpEchoEndpoint == "" {
		return "", errors.New("TCP_ECHO_ENDPOINT is not set")
	}
	return tcpEchoEndpoint, nil
}

func (h *colorHandler) TCPEcho(ctx context.Context, empty *frontservice.Empty) (*frontservice.Empty, error) {
	return &frontservice.Empty{}, nil
}

func (h *colorHandler) Ping(ctx context.Context, empty *frontservice.Empty) (*frontservice.Empty, error) {
	return &frontservice.Empty{}, nil
}

func (h *colorHandler) ClearColor(ctx context.Context, empty *frontservice.Empty) (*frontservice.Message, error) {
	colorsMutext.Lock()
	defer colorsMutext.Unlock()

	colorsIdx = 0
	for i := range colors {
		colors[i] = ""
	}

	return &frontservice.Message{Color: "cleared"}, nil
}

func main() {
	log.Println("Starting server, listening on port " + getServerPort())

	colorTellerEndpoint, err := getColorTellerEndpoint()
	if err != nil {
		log.Fatalln(err)
	}
	tcpEchoEndpoint, err := getTCPEchoEndpoint()
	if err != nil {
		log.Println(err)
	}

	log.Println("Using color-teller at " + colorTellerEndpoint)
	log.Println("Using tcp-echo at " + tcpEchoEndpoint)

	xraySegmentNamer := xray.NewFixedSegmentNamer(fmt.Sprintf("%s-gateway", getStage()))
	server := NewColorHandler()
	twirpHandler := frontservice.NewGatewayServiceServer(server, NewXrayServerHooks())
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
