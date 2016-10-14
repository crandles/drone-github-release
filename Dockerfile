FROM alpine:latest

RUN apk add --update subversion curl && rm -rf /var/cache/apk/*

ADD drone-svn-release /bin/

ENTRYPOINT ["/bin/drone-svn-release"]
