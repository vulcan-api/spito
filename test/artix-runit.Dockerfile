FROM artixlinux/artixlinux:base-runit

RUN pacman -Sy go --noconfirm

WORKDIR /app
COPY .. .
USER 0

RUN go build

CMD ["/sbin/init"]
