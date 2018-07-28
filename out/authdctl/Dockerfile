FROM alpine:3.8

ADD ./authdctl /bin/authdctl

RUN adduser -D -s /sbin/nologin authdctl

USER authdctl

ENTRYPOINT [ "authdctl" ]
