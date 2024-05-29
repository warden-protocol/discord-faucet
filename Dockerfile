FROM ubuntu:22.04 as wardend

RUN apt-get update && apt-get install -y \
    ca-certificates \
    wget \
    unzip \
    && rm -rf /var/lib/apt/lists/*

RUN wget https://github.com/warden-protocol/wardenprotocol/releases/download/v0.3.0/wardend_Linux_x86_64.zip -O /tmp/wardend.zip \
    && unzip /tmp/wardend.zip -d ./ \
    && rm /tmp/wardend.zip

#FROM gcr.io/distroless/base-debian12:nonroot
FROM ubuntu:22.04

RUN apt-get update && apt-get install -y \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY discord-faucet /usr/bin/discord-faucet
COPY --from=wardend /wardend /usr/bin/wardend

CMD ["/usr/bin/discord-faucet"]
#ENTRYPOINT ["ls", "-la", "/usr/bin"]
