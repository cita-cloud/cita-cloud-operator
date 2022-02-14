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

package exec

import (
	"testing"
)

func TestInitChain(t *testing.T) {
	cc := NewCloudConfig("my-chain", "./")
	_ = cc.Init("11111111111")
}

func TestCreateAccount(t *testing.T) {
	cc := NewCloudConfig("my-chain", "./")
	_, address, err := cc.CreateAccount("123456")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(address)
}

func TestCreateCa(t *testing.T) {
	cc := NewCloudConfig("my-chain", "./")
	_, _, _ = cc.CreateCaAndRead()
}

func TestDeleteChain(t *testing.T) {
	cc := NewCloudConfig("my-chain", "./")
	_ = cc.DeleteChain()
}
