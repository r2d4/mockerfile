
FROM golang:1.11-alpine AS builder
COPY . /go/src/github.com/r2d4/mockerfile
RUN CGO_ENABLED=0 go build -o /mocker -tags "$BUILDTAGS" --ldflags '-extldflags "-static"' github.com/r2d4/mockerfile/cmd/mocker

FROM scratch
COPY --from=builder /mocker /bin/mocker
ENTRYPOINT ["/bin/mocker"]