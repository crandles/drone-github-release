FROM alpine:latest

RUN apk add --update subversion curl openssl-dev

ADD drone-svn-release /bin/

ENTRYPOINT ["/bin/drone-svn-release"]
