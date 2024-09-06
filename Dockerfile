FROM archlinux AS base

RUN pacman -Syu --noconfirm

RUN pacman -S --noconfirm go

FROM base AS build

ENV PORT=$PORT

COPY . /app
WORKDIR /app

ENTRYPOINT ["go", "run", "main.go"]
