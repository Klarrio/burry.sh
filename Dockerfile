FROM alpine
MAINTAINER Tim Wuyts <tim.wuyts@klarrio.com>
RUN apk --update upgrade && apk add ca-certificates && apk add libc6-compat && \
    addgroup -g 1001 dsh && adduser -u 1001 -D -H -G dsh dsh && \
    mkdir /app && chown dsh:dsh app
RUN apk add --no-cache tini
ENTRYPOINT ["/sbin/tini", "--"]
USER dsh
ADD burry.sh continuous.sh /app/
CMD ["/app/continuous.sh"]
