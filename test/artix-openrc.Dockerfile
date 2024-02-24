FROM artixlinux/artixlinux:base-openrc

RUN pacman -Sy go --noconfirm

WORKDIR /app
COPY .. .

USER root
RUN go build

CMD ["/sbin/init"]
