FROM golang:1.12 as build

RUN go get github.com/docker/docker/api \
    && go get github.com/docker/docker/client

WORKDIR /go/src/rmazur.io/engine-check
COPY . .
RUN go install -ldflags "-extldflags '-fno-PIC -static'" -buildmode pie -tags 'osusergo netgo static_build' rmazur.io/engine-check/cmd/engine-check

FROM scratch
COPY --from=build /go/bin/engine-check /engine-check
ENTRYPOINT ["/engine-check"]
