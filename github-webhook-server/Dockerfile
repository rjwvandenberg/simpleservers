FROM golang:1.22.2 as build
WORKDIR /build
COPY github-webhook-server/ /build
RUN go mod download
RUN CGO_ENABLED=0 go build -o artifacts/server

# prod container
FROM gcr.io/distroless/static:latest
WORKDIR /bin
COPY --from=build /build/artifacts/server /bin/
CMD ["/bin/server"]
