FROM golang:1.15 AS build
RUN groupadd --non-unique --gid 1001 build-group \
    && useradd --non-unique -m --uid 1001 --gid 1001 build-user

RUN mkdir /build
WORKDIR /build
COPY go.mod go.sum /build/
RUN pwd
RUN ls -l
RUN go list -m all
RUN go mod download

ADD . /build
RUN make


FROM gcr.io/distroless/static
USER nonroot
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


