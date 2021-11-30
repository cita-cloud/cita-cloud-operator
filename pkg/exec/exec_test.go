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
