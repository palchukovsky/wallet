ARG NODE_OS_NAME
ARG NODE_OS_TAG
ARG BUILDER

FROM ${BUILDER} as builder
RUN go test -timeout 10s ./cmd/rest-server -v
RUN go install -v ./cmd/rest-server 

FROM ${NODE_OS_NAME}:${NODE_OS_TAG}
WORKDIR /opt/palchukovsky/
COPY --from=builder /go/bin/rest-server .
ENTRYPOINT ["./rest-server"]