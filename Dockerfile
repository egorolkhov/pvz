FROM ubuntu:latest
LABEL authors="egoro"

ENTRYPOINT ["top", "-b"]