#!/bin/bash
#
# Copyright Rivtower Technologies LLC.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
#

set -o errexit

echo "create namespace cita..."
kubectl create ns cita
echo "create namespace cita successful!"

echo "install cita-cloud-operator..."
helm repo add cita-cloud-operator https://cita-cloud.github.io/cita-cloud-operator
helm install cita-cloud-operator cita-cloud-operator/cita-cloud-operator -n=cita

times=300
while [ $times -ge 1 ]
do
  if [ `kubectl get pod -ncita | grep cita-cloud-operator | awk '{print $3}'` == "Running" ]; then
    break
  else
    echo "cita-cloud-operator pod is not Running..."
    let times--
    sleep 1
  fi
done
if [ $times -lt 1 ]; then
  echo "wait timeout for cita-cloud-operator"
  exit 1
else
  echo "cita-cloud-operator is Running"
fi

echo "install operator-proxy..."
helm repo add cita-cloud-operator-proxy https://cita-cloud.github.io/operator-proxy
helm install cita-cloud-operator-proxy cita-cloud-operator-proxy/cita-cloud-operator-proxy -n=cita

times=300
while [ $times -ge 1 ]
do
  if [ `kubectl get pod -ncita | grep cita-cloud-operator-proxy | awk '{print $3}'` == "Running" ];then
    break
  else
    echo "cita-cloud-operator-proxy pod is not Running..."
    let times--
    sleep 1
  fi
done
if [ $times -lt 1 ]; then
  echo "wait timeout for cita-cloud-operator-proxy"
  exit 1
else
  echo "cita-cloud-operator-proxy is Running"
fi

endpoint=`kubectl get  svc cita-cloud-operator-proxy -ncita -ojson | jq '.spec.ports[0].nodePort'`

curl -sSL https://github.com/cita-cloud/operator-proxy/releases/download/v0.0.1-alpha/cco-cli-0.0.1-alpha-linux-amd64 --output /tmp/cco-cli
chmod +x /tmp/cco-cli
sudo mv /tmp/cco-cli /usr/local/bin
export OPERATOR_PROXY_ENDPOINT=127.0.0.1:$endpoint
cco-cli -h
telnet 127.0.0.1 $endpoint