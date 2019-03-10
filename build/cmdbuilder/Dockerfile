ARG TAG

FROM golang:${TAG}

WORKDIR /go/src/github.com/palchukovsky/wallet
COPY . .

RUN apk update
RUN apk add git
RUN apk add build-base

RUN go get -v github.com/lib/pq
RUN go get -v github.com/gorilla/mux
RUN go get -v github.com/golang/mock/gomock
RUN go get -v github.com/golang/mock/mockgen
RUN make mock
RUN go test -timeout 15s ./ -v
