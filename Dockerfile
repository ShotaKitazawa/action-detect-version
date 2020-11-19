FROM alpine/git:latest

RUN apk add --no-cache bash curl jq

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
