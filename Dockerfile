FROM archlinux AS base

RUN pacman -Syu --noconfirm

ENTRYPOINT ["tail", "-f", "/dev/null"]
