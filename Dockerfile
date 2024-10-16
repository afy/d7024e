FROM archlinux AS base

RUN pacman -Syu --noconfirm

RUN pacman -S --noconfirm go

FROM base AS build

WORKDIR /app
COPY ./go.mod /app/go.mod

RUN ["go", "mod", "download"]

COPY ./cli.sh /usr/bin/kad

ENV PORT=$PORT
ENV IS_BOOTSTRAP_NODE=$IS_BOOTSTRAP_NODE

COPY . /app


ENTRYPOINT ["go", "run", "main.go"]
