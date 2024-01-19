FROM artixlinux/openrc

RUN pacman -Sy go git --noconfirm

WORKDIR /app
COPY .. .

RUN git submodule update --init
RUN go build

CMD ["/sbin/init"]
