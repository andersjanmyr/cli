FROM rust:latest

ENV FASTLY_CLI_VERSION 0.27.1

WORKDIR /tmp
COPY dockerfiles/* dockerfiles/.cargo ./
RUN fastly compute build || true
RUN rm -rf /tmp/* /tmp/.cargo

WORKDIR /app
ENTRYPOINT ["fastly"]
CMD ["--help"]
