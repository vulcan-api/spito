FROM archlinux

RUN pacman -Sy go --noconfirm

WORKDIR /app
COPY ./src/ .

RUN go build

CMD ["go", "run", "."]
