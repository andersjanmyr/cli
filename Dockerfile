FROM rust:latest

WORKDIR /tmp
COPY dockerfiles/* dockerfiles/.cargo ./
RUN fastly compute build || true
RUN rm -rf /tmp/* /tmp/.cargo

WORKDIR /app
ENTRYPOINT ["fastly"]
CMD ["--help"]
