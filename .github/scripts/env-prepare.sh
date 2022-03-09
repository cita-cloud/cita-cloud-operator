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

#times=300
#while [ $times -ge 1 ]
#do
#  if [ `kubectl get pod -ncita | grep cita-cloud-operator | awk '{print $3}'` == "Running" ];then
#    break
#  else
#    echo "cita-cloud-operator pod is not Running..."
#    let times--
#    sleep 1
#  fi
#done
#if [ $times -lt 1 ]; then
#  echo "wait timeout for cita-cloud-operator"
#  exit 1
#else
#  echo "cita-cloud-operator is Running"
#fi
#
#times=300
#while [ $times -ge 1 ]
#do
#  if [ `kubectl get pod -ncita | grep cita-cloud-operator-proxy | awk '{print $3}'` == "Running" ];then
#    break
#  else
#    echo "cita-cloud-operator-proxy pod is not Running..."
#    let times--
#    sleep 1
#  fi
#done
#if [ $times -lt 1 ]; then
#  echo "wait timeout for cita-cloud-operator-proxy"
#  exit 1
#else
#  echo "cita-cloud-operator-proxy is Running"
#fi
#
#endpoint=`kubectl get  svc cita-cloud-operator-proxy -ncita -ojson | jq '.spec.ports[0].nodePort'`

export OWNER=cita-cloud
export REPO=operator-proxy
export BIN_LOCATION="/usr/local/bin"

cli_version=$(curl -s https://api.github.com/repos/$OWNER/$REPO/releases/latest | grep 'tag_name' | cut -d '"' -f 4 | tr -d 'v')

getPackage() {
    uname=$(uname)
    userid=$(id -u)

    suffix=""
    case $uname in
    "Darwin")
        arch=$(uname -m)
        case $arch in
        "x86_64")
        suffix="darwin-adm64"
        ;;
        esac
        case $arch in
        "arm64")
        suffix="darwin-arm64"
        ;;
        esac
    ;;

    "MINGW"*)
    suffix=".exe"
    BINLOCATION="$HOME/bin"
    mkdir -p $BINLOCATION

    ;;
    "Linux")
        arch=$(uname -m)
        echo $arch
        case $arch in
        "x86_64")
        suffix="linux-amd64"
        ;;
        esac
        case $arch in
        "aarch64")
        suffix="linux-arm64"
        ;;
        esac
    ;;
    esac

    targetFile="/tmp/cco-cli-$cli_version-$suffix.tar.gz"

    if [ -e "$targetFile" ]; then
        rm "$targetFile"
    fi

    url=https://github.com/$OWNER/$REPO/releases/download/v$cli_version/cco-cli-$cli_version-$suffix.tar.gz
    echo "Downloading package $url as $targetFile"

    curl -sSL $url --output "$targetFile"

    if [ "$?" = "0" ]; then
      echo "Download complete."
      tar -zxf $targetFile
      chmod +x cco-cli
      mv cco-cli $BIN_LOCATION
    fi
}

getPackage