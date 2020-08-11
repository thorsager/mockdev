FROM golang:1.14 AS builder
RUN groupadd --non-unique --gid 1001 build-group \
    && useradd --non-unique -m --uid 1001 --gid 1001 build-user

RUN mkdir /build && chown build-user /build
USER build-user
WORKDIR build

COPY go.mod go.sum /build/
RUN go mod download

ADD . /build
RUN make


FROM gcr.io/distroless/static
USER nonroot
WORKDIR /

COPY --from=build /build/gollo /

EXPOSE 8080

ENTRYPOINT [ "/gollo" ]

