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
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	utilexec "k8s.io/utils/exec"
)

const cloudconfig = "cloud-config"

type Cmd interface {
	Exist() bool
	Init(chainId string) error
	CreateAccount(kmsPassword string) (keyId, address string, err error)
	ReadKmsDb(address string) ([]byte, error)
	// CreateCaAndRead. if no errors, return content of cert.pem，key.pem
	CreateCaAndRead() ([]byte, []byte, error)
	CaExist() bool
	ReadCa() ([]byte, []byte, error)
	WriteCaCert(cert []byte) error
	WriteCaKey(key []byte) error
	CreateSignCsrAndRead(domain string) ([]byte, []byte, []byte, error)
	//ReadCsr(domain string) ([]byte, []byte, []byte, error)
	DeleteChain() error
}

type CloudConfigCmd struct {
	chainName string
	configDir string
	exec      utilexec.Interface
}

func (c CloudConfigCmd) Exist() bool {
	_, err := os.Stat(fmt.Sprintf("%s/%s", c.configDir, c.chainName))
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

func (c CloudConfigCmd) Init(chainId string) error {
	err := c.exec.Command(cloudconfig, "init-chain", "--chain-name", c.chainName, "--config-dir", c.configDir).Run()
	if err != nil {
		return err
	}
	err = c.exec.Command(cloudconfig, "init-chain-config", "--chain-name", c.chainName, "--config-dir", c.configDir, "--chain_id", chainId).Run()
	if err != nil {
		return err
	}
	return nil
}

func (c CloudConfigCmd) CreateAccount(kmsPassword string) (keyId, address string, err error) {
	out, err := c.exec.Command(cloudconfig, "new-account", "--chain-name", c.chainName, "--config-dir", c.configDir, "--kms-password", kmsPassword).Output()
	if err != nil {
		return "", "", err
	}
	l := strings.Split(string(out), ",")
	return strings.Split(l[0], ":")[1], strings.Replace(strings.Split(l[1], ":")[1], "\n", "", -1), nil
}

func (c CloudConfigCmd) ReadKmsDb(address string) ([]byte, error) {
	buf, err := ioutil.ReadFile(fmt.Sprintf("%s/%s/accounts/%s/kms.db", c.configDir, c.chainName, address))
	if err != nil {
		return nil, err
	}

	return buf, nil
}

// CreateCaAndRead. if no errors, return content of cert.pem，key.pem
func (c CloudConfigCmd) CreateCaAndRead() ([]byte, []byte, error) {
	err := c.exec.Command(cloudconfig, "create-ca", "--chain-name", c.chainName, "--config-dir", c.configDir).Run()
	if err != nil {
		return nil, nil, err
	}
	cert, err := ioutil.ReadFile(fmt.Sprintf("%s/%s/ca_cert/cert.pem", c.configDir, c.chainName))
	if err != nil {
		return nil, nil, fmt.Errorf("read ca cert.pem error: %v", err)
	}
	key, err := ioutil.ReadFile(fmt.Sprintf("%s/%s/ca_cert/key.pem", c.configDir, c.chainName))
	if err != nil {
		return nil, nil, fmt.Errorf("read ca key.pem error: %v", err)
	}
	return cert, key, nil
}

func (c CloudConfigCmd) ReadCa() ([]byte, []byte, error) {
	cert, err := ioutil.ReadFile(fmt.Sprintf("%s/%s/ca_cert/cert.pem", c.configDir, c.chainName))
	if err != nil {
		return nil, nil, fmt.Errorf("read ca cert.pem error: %v", err)
	}
	key, err := ioutil.ReadFile(fmt.Sprintf("%s/%s/ca_cert/key.pem", c.configDir, c.chainName))
	if err != nil {
		return nil, nil, fmt.Errorf("read ca key.pem error: %v", err)
	}
	return cert, key, nil
}

func (c CloudConfigCmd) CaExist() bool {
	_, err := os.Stat(fmt.Sprintf("%s/%s/ca_cert/cert.pem", c.configDir, c.chainName))
	if err != nil {
		return false
	}
	return true
}

func (c CloudConfigCmd) WriteCaCert(cert []byte) error {
	err := ioutil.WriteFile(fmt.Sprintf("%s/%s/ca_cert/cert.pem", c.configDir, c.chainName), cert, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (c CloudConfigCmd) WriteCaKey(key []byte) error {
	err := ioutil.WriteFile(fmt.Sprintf("%s/%s/ca_cert/key.pem", c.configDir, c.chainName), key, 0644)
	if err != nil {
		return err
	}
	return nil
}

// CreateCsrAndRead. if no errors, return content of domain's csr.pem, key.pem, cert.pem
func (c CloudConfigCmd) CreateSignCsrAndRead(domain string) ([]byte, []byte, []byte, error) {
	err := c.exec.Command(cloudconfig, "create-csr", "--chain-name", c.chainName, "--config-dir", c.configDir, "--domain", domain).Run()
	if err != nil {
		return nil, nil, nil, err
	}
	csr, err := ioutil.ReadFile(fmt.Sprintf("%s/%s/certs/%s/csr.pem", c.configDir, c.chainName, domain))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("read domain csr.pem error: %v", err)
	}
	key, err := ioutil.ReadFile(fmt.Sprintf("%s/%s/certs/%s/key.pem", c.configDir, c.chainName, domain))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("read domain key.pem error: %v", err)
	}

	err = c.exec.Command(cloudconfig, "sign-csr", "--chain-name", c.chainName, "--config-dir", c.configDir, "--domain", domain).Run()
	if err != nil {
		return nil, nil, nil, err
	}
	cert, err := ioutil.ReadFile(fmt.Sprintf("%s/%s/certs/%s/cert.pem", c.configDir, c.chainName, domain))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("read domain cert.pem error: %v", err)
	}

	return csr, key, cert, nil
}

func (c CloudConfigCmd) DeleteChain() error {
	err := c.exec.Command(cloudconfig, "delete-chain", "--chain-name", c.chainName, "--config-dir", c.configDir).Run()
	if err != nil {
		return err
	}
	return nil
}

func NewCloudConfig(chainName string, configDir string) Cmd {
	return CloudConfigCmd{
		chainName: chainName,
		configDir: configDir,
		exec:      utilexec.New(),
	}
}
