FROM alpine:latest

RUN apk add --update subversion curl openssl-dev && rm -rf /var/cache/apk/*

ADD drone-svn-release /bin/
ADD make_svn_dir.sh /bin/
ADD push.sh /bin/

ENTRYPOINT ["/bin/drone-svn-release"]
