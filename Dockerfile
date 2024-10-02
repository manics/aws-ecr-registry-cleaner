# https://www.docker.com/blog/faster-multi-platform-builds-dockerfile-cross-compilation-guide/

###########################################################################
FROM --platform=$BUILDPLATFORM docker.io/library/golang:1.21-alpine AS build

RUN apk add --no-cache make git

WORKDIR /src
COPY . ./

ARG TARGETOS TARGETARCH
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH make build

###########################################################################
FROM docker.io/library/alpine:3.20

COPY --from=build /src/aws-ecr-registry-cleaner /bin/

RUN adduser -S -D -H -h /app appuser
USER appuser

ENTRYPOINT [ "aws-ecr-registry-cleaner" ]

EXPOSE 8080
