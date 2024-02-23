FROM archlinux

RUN pacman -Sy go --noconfirm

WORKDIR /app
COPY .. .
USER 0

RUN go build

CMD ["/sbin/init"]
