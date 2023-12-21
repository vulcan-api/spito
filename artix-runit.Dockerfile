FROM artixlinux/runit

RUN pacman -Sy go --noconfirm

WORKDIR /app
COPY ./src/ .

RUN go build

CMD ["/sbin/init"]
