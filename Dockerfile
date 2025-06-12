FROM registry.access.redhat.com/ubi9/go-toolset:1.23.9-1749636489 AS builder
COPY api api
COPY pkg pkg
COPY go.mod go.mod
COPY go.sum go.sum
COPY main.go main.go
COPY openapi.yaml openapi.yaml
COPY server.cfg.yaml server.cfg.yaml
COPY tools.go tools.go
COPY Makefile Makefile
USER root
RUN go get -d -v
RUN make generate
RUN CGO_ENABLED=1 go build -o /go/bin/widget-layout-backend 

FROM registry.access.redhat.com/ubi9-minimal:latest

# Setup permissions to allow RDSCA to be written from clowder to container
# https://docs.openshift.com/container-platform/4.11/openshift_images/create-images.html#images-create-guide-openshift_create-images
RUN mkdir -p /app
RUN chgrp -R 0 /app && \
    chmod -R g=u /app
COPY --from=builder /go/bin/widget-layout-backend /app/widget-layout-backend
# Spec is used for request payload validation
COPY openapi.yaml /app/openapi.yaml

WORKDIR /app
ENTRYPOINT ["./widget-layout-backend"]
EXPOSE 8000
USER 1001
