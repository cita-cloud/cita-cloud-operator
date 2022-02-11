package controllers

import (
	"github.com/pelletier/go-toml"
	"reflect"
	"testing"
)

func TestCheckNetworkConfigChanged(t *testing.T) {
	config, _ := toml.Load(`
[network_p2p]
grpc_port = 50000
port = 40000

[[network_p2p.peers]]
address = '/dns4/my-chain-a727a61063ab-nodeport/tcp/40000'

[[network_p2p.peers]]
address = '/dns4/my-chain-e15e0f10a97c-nodeport/tcp/40000'`)
	networkP2p := config.Get("network_p2p").(*toml.Tree)
	peers := networkP2p.GetArray("peers")
	if reflect.TypeOf(peers).Kind() == reflect.Slice {
		s := reflect.ValueOf(peers)
		t.Log(s.Len())
	}
	t.Log(peers)
}
