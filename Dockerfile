FROM golang:1.16-alpine as builder
RUN mkdir /mybuild
ADD . /mybuild/
WORKDIR /mybuild/cmd/telefacts-sec
RUN apk update && apk add --no-cache git
RUN go get -d -v
RUN CGO_ENABLED=0 GOOS=linux go build -o /mybuild/main /mybuild/cmd/telefacts-sec/main.go

FROM ghcr.io/ecksbee/goldlord-midas:main as spa

FROM ghcr.io/ecksbee/sec-testdata:main as secdata

FROM alpine:latest
RUN apk --update add ca-certificates
COPY --from=secdata /wd /wd
COPY --from=secdata /gts /gts
COPY --from=builder /mybuild/main /
COPY --from=builder /mybuild/cmd/telefacts-sec/filing.tmpl /
COPY --from=builder /mybuild/cmd/telefacts-sec/home.tmpl /
COPY --from=builder /mybuild/cmd/telefacts-sec/import.tmpl /
COPY --from=builder /mybuild/cmd/telefacts-sec/search.tmpl /
COPY --from=spa / /wd/goldlord-midas
WORKDIR /
RUN chown -R 1000:1000 /wd
RUN chown -R 1000:1000 /gts
USER 1000
EXPOSE 8080
ENTRYPOINT ["./main"]