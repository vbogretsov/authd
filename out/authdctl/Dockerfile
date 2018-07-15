FROM alpine:3.7

ENV DBURL=

ADD ./authdctl /bin/authdctl

RUN adduser -D authdctl

USER authdctl

ENTRYPOINT [ "authdctl" ]
