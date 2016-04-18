FROM debian:jessie

RUN apt-get update && \
  apt-get upgrade -y && \
  apt-get install -y git && \
  apt-get clean -y && \
  apt-get autoclean -y && \
  apt-get autoremove -y && \
  mkdir -p /var/log/postgres-ci/ && \
  rm -rf /usr/share/locale/* && \
  rm -rf /var/cache/debconf/*-old && \
  rm -rf /var/lib/apt/lists/* && \
  rm -rf /usr/share/doc/*

COPY ./tmp/worker /usr/local/bin/postgres-ci-worker
ENTRYPOINT ["/usr/local/bin/postgres-ci-worker"]