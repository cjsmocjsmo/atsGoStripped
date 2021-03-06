FROM golang:bullseye AS builder
RUN mkdir /go/src/atsGo
WORKDIR /go/src/atsGo

COPY atsGo.go .

COPY go.mod .
COPY go.sum .
RUN export GOPATH=/go/src/atsGo
RUN go get -v /go/src/atsGo
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main /go/src/atsGo

# FROM alpine:latest
FROM debian:bullseye
WORKDIR /root/

COPY --from=builder /go/src/atsGo/main .

RUN \
  mkdir ./data && \
  mkdir ./data/db && \
  mkdir ./static && \
  mkdir ./backup && \
  mkdir ./static/images && \
  mkdir ./fsData && \
  mkdir ./fsData/thumb && \
  mkdir ./fsData/crap && \
  mkdir ./logs && \
  mkdir ./certs

COPY backup/*.json ./backup/
COPY backup/*.gz ./backup/
COPY static/*.html ./static/
COPY static/*.css ./static/
COPY static/*.js ./static/
COPY static/*.yaml ./static/
COPY static/images/*.jpg ./static/images/
COPY static/images/*.png ./static/images

RUN \
  snap install core && \ 
  snap refresh core && \
  snap install --classic certbot && \
  ln -s /snap/bin/certbot /usr/bin/certbot && \
  certbot certonly --standalone --agree-tos -m charlie@atsio.xyz -d atsio.xyz --cert-path ./certs --key-path ./certs && \
  chmod -R +rwx ./static && \
  chmod -R +rwx ./fsData && \
  chmod -R +rwx ./logs && \
  chmod -R +rwx ./backup

STOPSIGNAL SIGINT
CMD ["./main"]

