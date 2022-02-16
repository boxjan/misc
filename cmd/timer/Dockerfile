FROM alpine

LABEL maintainers="boxjan"
LABEL description="what time is it now?"

ARG binary=./bin/timer
ARG os=""
ARG arch=""

COPY ${binary}_${os}_${arch} timer
ENTRYPOINT ["/timer"]