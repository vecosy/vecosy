FROM golang:latest as builder
ARG VCS_REF
LABEL org.label-schema.build-date=$BUILD_DATE \
      org.label-schema.vcs-ref=$VCS_REF \
      org.label-schema.vcs-url="https://github.com/vecosy/vecosy"
WORKDIR /go/src
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o vecosy-server

FROM alpine:latest
RUN mkdir /config
EXPOSE 8080
EXPOSE 8081
COPY --from=builder /go/src/vecosy-server /vecosy-server
CMD /vecosy-server --config /config/vecosy.yml
