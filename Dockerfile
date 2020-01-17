FROM alpine
MAINTAINER Tim Wuyts <tim.wuyts@klarrio.com>
RUN apk --update upgrade && apk add ca-certificates && apk add libc6-compat && \
    addgroup -g 1001 dsh && adduser -u 1001 -D -H -G dsh dsh && \
    mkdir /app && chown dsh:dsh app
USER dsh
ADD burry.sh daemon.sh /app/
CMD ["/app/daemon.sh"]
