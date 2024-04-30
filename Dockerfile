FROM golang:1.22.1-alpine3.19 as build

WORKDIR /app
COPY . /app
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-extldflags=-static -X github.com/prodyna/delete-from-enterprise/meta.Version=1.0.0" -o delete-from-enterprise main.go

FROM alpine:3.19.1
COPY --from=build /app/delete-from-enterprise /app/
# COPY /template /template
ENTRYPOINT ["/app/delete-from-enterprise"]
