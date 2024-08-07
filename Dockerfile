FROM golang:1.22-alpine as builder
RUN mkdir /mybuild
ADD . /mybuild/
WORKDIR /mybuild/cmd/telefacts-sec
RUN apk update && apk add --no-cache git
RUN CGO_ENABLED=0 GOOS=linux go build -o /mybuild/main /mybuild/cmd/telefacts-sec/main.go

FROM ghcr.io/ecksbee/goldlord-midas:main as spa

FROM ghcr.io/ecksbee/snakebane-patrick:main as ssg

FROM ghcr.io/ecksbee/sec-testdata:main as secdata

FROM alpine:latest
RUN apk --update add ca-certificates
COPY --from=secdata /wd /wd
COPY --from=secdata /gts /gts
COPY --from=builder /mybuild/main /
COPY --from=spa / /goldlord-midas
COPY --from=ssg / /snakebane-patrick
WORKDIR /
RUN chown -R 1000:1000 /wd
RUN chown -R 1000:1000 /gts
RUN chown -R 1000:1000 /goldlord-midas
RUN chown -R 1000:1000 /snakebane-patrick
USER 1000
EXPOSE 8080
ENTRYPOINT ["./main"]