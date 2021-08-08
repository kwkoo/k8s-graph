FROM golang:1.16.6 as builder
ARG PREFIX=github.com/kwkoo
ARG PACKAGE=k8s-graph
LABEL builder=true
LABEL org.opencontainers.image.source https://github.com/kwkoo/k8s-graph
COPY . /go/src/
RUN \
  set -x \
  && \
  cd /go/src/ \
  && \
  CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /go/bin/${PACKAGE}

FROM scratch
LABEL maintainer="kin.wai.koo@gmail.com"
LABEL builder=false

EXPOSE 8080
USER 1001
ENTRYPOINT ["/k8s-graph"]

# we need to copy the certificates over because we're connecting over SSL
COPY --from=builder /etc/ssl /etc/ssl

COPY --from=builder /go/bin/${PACKAGE} /
