FROM archlinux AS base

RUN pacman -Syu --noconfirm

RUN pacman -S --noconfirm go

FROM base AS build

WORKDIR /app
COPY ./go.mod /app/go.mod

RUN ["go", "mod", "download"]

ENV PORT=$PORT

COPY . /app


ENTRYPOINT ["go", "run", "main.go"]
