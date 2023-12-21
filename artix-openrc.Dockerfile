FROM artixlinux/openrc

RUN pacman -Sy go --noconfirm

WORKDIR /app
COPY ./src/ .

RUN go build

CMD ["/sbin/init"]
