FROM ubuntu:22.04 as wardend

RUN apt-get update && apt-get install -y \
  ca-certificates \
  wget \
  unzip \
  && rm -rf /var/lib/apt/lists/*

RUN wget https://github.com/warden-protocol/wardenprotocol/releases/download/v0.5.3/wardend_Linux_x86_64.zip -O /tmp/wardend.zip \
  && unzip /tmp/wardend.zip -d ./ \
  && rm /tmp/wardend.zip

FROM ubuntu:22.04

RUN apt-get update && apt-get install -y \
  ca-certificates \
  && rm -rf /var/lib/apt/lists/*

RUN groupadd -r warden && useradd --no-log-init -r -g warden warden

USER warden

WORKDIR /app

COPY discord-faucet /app/discord-faucet
COPY --from=wardend /tmp/wardend /app/wardend

CMD ["/usr/bin/discord-faucet"]
