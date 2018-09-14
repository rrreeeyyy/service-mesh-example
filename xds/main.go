package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"github.com/envoyproxy/go-control-plane/pkg/cache"
	xds "github.com/envoyproxy/go-control-plane/pkg/server"

	log "github.com/sirupsen/logrus"
)

const (
	nodeId      = "front-proxy"
	gatewayPort = uint(80)
)

func main() {
	ctx := context.Background()

	config := cache.NewSnapshotCache(false, Hasher{}, logger{})

	signal := make(chan struct{})
	cb := &callbacks{signal: signal}

	snapshot := GenerateExampleSnapshot()

	srv := xds.NewServer(config, cb)

	go RunManagementGateway(ctx, srv, gatewayPort)

	err := config.SetSnapshot(nodeId, snapshot)
	if err != nil {
		log.Errorf("snapshot error %q for %+v", err, snapshot)
		os.Exit(1)
	}

	<-signal
}

func RunManagementGateway(ctx context.Context, srv xds.Server, port uint) {
	log.WithFields(log.Fields{"port": port}).Info("gateway listening HTTP/1.1")
	server := &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: &xds.HTTPGateway{Server: srv}}
	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Error(err)
		}
	}()
}

type Hasher struct {
}

func (h Hasher) ID(node *core.Node) string {
	if node == nil {
		return "unknown"
	}
	return node.Id
}

type logger struct{}

func (logger logger) Infof(format string, args ...interface{}) {
	log.Debugf(format, args...)
}
func (logger logger) Errorf(format string, args ...interface{}) {
	log.Errorf(format, args...)
}

type callbacks struct {
	signal   chan struct{}
	fetches  int
	requests int
	mu       sync.Mutex
}

func (cb *callbacks) Report()                                                             {}
func (cb *callbacks) OnStreamOpen(id int64, typ string)                                   {}
func (cb *callbacks) OnStreamClosed(id int64)                                             {}
func (cb *callbacks) OnStreamRequest(int64, *v2.DiscoveryRequest)                         {}
func (cb *callbacks) OnStreamResponse(int64, *v2.DiscoveryRequest, *v2.DiscoveryResponse) {}
func (cb *callbacks) OnFetchRequest(req *v2.DiscoveryRequest)                             {}
func (cb *callbacks) OnFetchResponse(*v2.DiscoveryRequest, *v2.DiscoveryResponse)         {}
