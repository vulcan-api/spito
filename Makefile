INSTALL_PREFIX=/usr/bin
all:
	go build .
install:
	cp spito $(INSTALL_PREFIX)
