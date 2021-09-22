package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/aws/aws-xray-sdk-go/xray"
	"golang.org/x/net/http2"

	"github.com/sirupsen/logrus"

	gologging "github.com/devopsfaith/krakend-gologging"
	viper "github.com/devopsfaith/krakend-viper"
	"github.com/gin-gonic/gin"
	_ "github.com/kyawmyintthein/aws-app-mesh-examples/colorapp/gateway/plugins/bodymodifier"
	_ "github.com/kyawmyintthein/aws-app-mesh-examples/colorapp/gateway/plugins/querystring2body"
	"github.com/kyawmyintthein/aws-app-mesh-examples/colorapp/gateway/rpc/colorteller"
	luratwirp "github.com/kyawmyintthein/lura-twirp"

	"github.com/luraproject/lura/proxy"
	krakendgin "github.com/luraproject/lura/router/gin"
	"github.com/luraproject/lura/transport/http/client"
	"github.com/luraproject/lura/transport/http/server"
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

func main() {
	port := flag.Int("p", 0, "Port of the service")
	debug := flag.Bool("d", false, "Enable the debug")
	configFile := flag.String("c", "./configuration.json", "Path to the configuration filename")
	flag.Parse()

	parser := viper.New()
	serviceConfig, err := parser.Parse(*configFile)
	if err != nil {
		log.Fatal("ERROR:", err.Error())
	}
	serviceConfig.Debug = serviceConfig.Debug || *debug
	if *port != 0 {
		serviceConfig.Port = *port
	}

	logger, err := gologging.NewLogger(serviceConfig.ExtraConfig, os.Stdout)
	if err != nil {
		log.Fatal("ERROR:", err.Error())
	}

	logger.Debug("config:", serviceConfig)

	ctx, cancel := context.WithCancel(context.Background())

	//backendFactory := martian.NewBackendFactory(logger, client.DefaultHTTPRequestExecutor(client.NewHTTPClient))
	colortellerLuraClient, err := colorteller.NewColortellerServiceLuraClient(&serviceConfig, "rpc.colorteller.ColortellerService", &http.Client{}, logger, twirp.WithClientPathPrefix("rz"))
	if err != nil {
		panic(err)
	}
	luratwirp.RegisterTwirpStubs(logger, colortellerLuraClient)

	bf := luratwirp.NewTwirpProxy(logger, client.DefaultHTTPRequestExecutor(client.NewHTTPClient))
	routerFactory := krakendgin.NewFactory(krakendgin.Config{
		Engine:         gin.Default(),
		Logger:         logger,
		Middlewares:    []gin.HandlerFunc{},
		HandlerFactory: krakendgin.EndpointHandler,
		ProxyFactory:   proxy.NewDefaultFactory(bf, logger),
		RunServer:      server.RunServer,
	})
	// go func() {
	// 	log.Println(http.ListenAndServe("localhost:6060", nil))
	// }()
	routerFactory.NewWithContext(ctx).Run(serviceConfig)

	cancel()
}

func transport2() *http2.Transport {
	return &http2.Transport{
		TLSClientConfig:    tlsConfig(),
		DisableCompression: true,
		AllowHTTP:          false,
	}
}

func tlsConfig() *tls.Config {
	crt, err := ioutil.ReadFile("./server.crt")
	if err != nil {
		log.Fatal(err)
	}

	rootCAs := x509.NewCertPool()
	rootCAs.AppendCertsFromPEM(crt)

	return &tls.Config{
		RootCAs:            rootCAs,
		InsecureSkipVerify: false,
		ServerName:         "localhost",
	}
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
