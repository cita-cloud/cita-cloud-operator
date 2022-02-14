/*
 * Copyright Rivtower Technologies LLC.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

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
