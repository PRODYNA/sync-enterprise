FROM golang:1.22.2-alpine3.19 as build

WORKDIR /app
COPY . .
RUN go build . -o delete-from-enterprise

FROM alpine:3.19.1
COPY --from=build /app/delete-from-enterprise /app/
# COPY /template /template
ENTRYPOINT ["/app/delete-from-enterprise"]
