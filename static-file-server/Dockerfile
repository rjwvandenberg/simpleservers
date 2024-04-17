# https://github.com/GoogleContainerTools/distroless
# https://github.com/GoogleContainerTools/distroless/blob/main/base/README.md
# build container
# assumes build context simpleservers/, not this subdirectory
FROM golang:1.22.2 as build
WORKDIR /build
COPY static-file-server/ /build
RUN go mod download
RUN CGO_ENABLED=0 go build -o artifacts/server

# prod container
FROM gcr.io/distroless/static:latest
WORKDIR /bin
COPY --from=build /build/artifacts/server /bin/
COPY testfiles/index.html /www/
COPY testfiles/index.html /www/static/
CMD ["/bin/server", "8080", "/www"]