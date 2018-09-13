package main

import (
	"net"
	"time"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/endpoint"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/route"
	"github.com/envoyproxy/go-control-plane/pkg/cache"
)

const (
	version    = "v1"
	localhost  = "127.0.0.1"
	XdsCluster = "xds_cluster"
)

var (
	RefreshDelay = 10 * time.Second
)

func MakeServicesRoute() *v2.RouteConfiguration {
	return &v2.RouteConfiguration{
		Name: "local_route",
		VirtualHosts: []route.VirtualHost{{
			Name:    "backend",
			Domains: []string{"*"},
			Routes: []route.Route{
				{
					Match: route.RouteMatch{
						PathSpecifier: &route.RouteMatch_Prefix{
							Prefix: "/service/1",
						},
					},
					Action: &route.Route_Route{
						Route: &route.RouteAction{
							ClusterSpecifier: &route.RouteAction_Cluster{
								Cluster: "service1",
							},
						},
					},
				},
				{
					Match: route.RouteMatch{
						PathSpecifier: &route.RouteMatch_Prefix{
							Prefix: "/service/2",
						},
					},
					Action: &route.Route_Route{
						Route: &route.RouteAction{
							ClusterSpecifier: &route.RouteAction_Cluster{
								Cluster: "service2",
							},
						},
					},
				},
			},
		}},
	}
}

func MakeEndpoint(clusterName string, port uint32) *v2.ClusterLoadAssignment {
	// Resolve IP address for https://github.com/envoyproxy/go-control-plane/issues/87
	addr, _ := net.ResolveIPAddr("ip4", clusterName)
	return &v2.ClusterLoadAssignment{
		ClusterName: clusterName,
		Endpoints: []endpoint.LocalityLbEndpoints{{
			LbEndpoints: []endpoint.LbEndpoint{{
				Endpoint: &endpoint.Endpoint{
					Address: &core.Address{
						Address: &core.Address_SocketAddress{
							SocketAddress: &core.SocketAddress{
								Protocol: core.TCP,
								Address:  addr.String(),
								PortSpecifier: &core.SocketAddress_PortValue{
									PortValue: port,
								},
							},
						},
					},
				},
			}},
		}},
	}
}

func MakeCluster(clusterName string) *v2.Cluster {
	return &v2.Cluster{
		Name:           clusterName,
		ConnectTimeout: 5 * time.Second,
		Type:           v2.Cluster_EDS,
		EdsClusterConfig: &v2.Cluster_EdsClusterConfig{
			EdsConfig: &core.ConfigSource{
				ConfigSourceSpecifier: &core.ConfigSource_ApiConfigSource{
					ApiConfigSource: &core.ApiConfigSource{
						ApiType:      core.ApiConfigSource_REST,
						ClusterNames: []string{XdsCluster},
						RefreshDelay: &RefreshDelay,
					},
				},
			},
		},
	}
}

func GenerateExampleSnapshot() cache.Snapshot {
	routes := make([]cache.Resource, 1)
	clusters := make([]cache.Resource, 2)
	endpoints := make([]cache.Resource, 2)

	routes[0] = MakeServicesRoute()
	clusters[0] = MakeCluster("service1")
	clusters[1] = MakeCluster("service2")

	endpoints[0] = MakeEndpoint("service1", 80)
	endpoints[1] = MakeEndpoint("service2", 80)

	return cache.NewSnapshot(version, endpoints, clusters, routes, nil)
}
