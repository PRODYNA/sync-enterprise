FROM golang:1.22.2-alpine3.19 as build

WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o delete-from-enterprise .

FROM alpine:3.19.1
COPY --from=build /app/delete-from-enterprise /app/
# COPY /template /template
ENTRYPOINT ["/app/delete-from-enterprise"]
