FROM registry.access.redhat.com/hi/go:latest-fips-builder AS builder
USER 0
WORKDIR /workspace

# Cache dependencies before copying source
COPY go.mod go.sum ./
RUN go mod download

COPY api api
COPY pkg pkg
COPY cmd cmd
COPY main.go main.go
COPY spec/openapi.yaml spec/openapi.yaml
COPY server.cfg.yaml server.cfg.yaml
COPY tools.go tools.go
COPY Makefile Makefile

RUN make generate
RUN CGO_ENABLED=1 go build -ldflags "-w -s" -o widget-layout-backend .
RUN CGO_ENABLED=1 go build -ldflags "-w -s" -o widget-layout-backend-migrate cmd/database/migrate.go

FROM registry.access.redhat.com/hi/go:latest-fips

# Setup permissions to allow RDSCA to be written from clowder to container
# https://docs.openshift.com/container-platform/4.11/openshift_images/create-images.html#images-create-guide-openshift_create-images
RUN mkdir -p /app/spec && \
    chgrp -R 0 /app && \
    chmod -R g=u /app

COPY --from=builder /workspace/widget-layout-backend /app/widget-layout-backend
COPY --from=builder /workspace/widget-layout-backend-migrate /usr/bin/widget-layout-backend-migrate
# Spec is used for request payload validation
COPY spec/openapi.yaml /app/spec/openapi.yaml

WORKDIR /app
EXPOSE 8000
USER 1001
CMD ["./widget-layout-backend"]
