FROM golang:1.11 AS build

ARG GO111MODULE=on

WORKDIR /go/src/github.com/irairdon/kustomize/internal/tools
COPY . /go/src/github.com/irairdon/kustomize/internal/tools

RUN go mod download
RUN CGO_ENABLED=0 go install github.com/irairdon/kustomize/internal/tools/cmd/backend/

FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /go/bin/backend /
ENTRYPOINT ["/backend"]
