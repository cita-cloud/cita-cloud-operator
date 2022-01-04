# Build the manager binary
FROM golang:1.16 as builder

WORKDIR /workspace/src/github.com/cita-cloud/cita-cloud-operator
ENV GOPATH=/workspace
ENV GO111MODULE=off
# Copy the Go Modules manifests
#COPY go.mod go.mod
#COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
#RUN go mod download

# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/
COPY pkg/ pkg/
COPY vendor/ vendor/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o cita-cloud-operator main.go

FROM debian:buster-slim
ARG version
ENV GIT_COMMIT=$version
LABEL maintainers="rivtower.com"
LABEL description="cita-cloud-operator"

MAINTAINER https://github.com/acechef

WORKDIR /
COPY --from=citacloud/cloud-config:v6.3.1 /usr/bin/cloud-config /usr/bin/
RUN /bin/sh -c set -eux;\
    apt-get update;\
    apt-get install -y --no-install-recommends libsqlite3-0;\
    rm -rf /var/lib/apt/lists/*;
COPY --from=builder /workspace/src/github.com/cita-cloud/cita-cloud-operator .
#USER 65532:65532
ENTRYPOINT ["/cita-cloud-operator"]
# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
#FROM gcr.io/distroless/static:nonroot
#WORKDIR /
#COPY --from=builder /workspace/manager .

#
#ENTRYPOINT ["/manager"]
