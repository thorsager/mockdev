FROM golang:1.15-alpine3.12 AS build
RUN apk add --update --no-cache make git
RUN mkdir /build
WORKDIR /build
COPY go.mod go.sum /build/
RUN go list -m all
RUN go mod download

ADD . /build
RUN git status
RUN make


FROM alpine:3.12
LABEL org.opencontainers.image.source=https://github.com/thorsager/mockdev
WORKDIR /
VOLUME /config

COPY --from=build /build/bin/mockdevd /
COPY --from=build /build/bin/snmp-snapshot /
COPY --from=build /build/bin/http-dump /
COPY resources/docker_default_config.yaml /config/mockdev.yaml

ENV PATH=/
EXPOSE 1161/udp
EXPOSE 8080/tcp
EXPOSE 2222/tcp

CMD [ "/mockdevd", "-c", "/config/mockdev.yaml" ]


