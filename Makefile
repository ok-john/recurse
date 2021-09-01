#
#	make - builds local binary.
#	make service - inits a systemd service for that binary.
#   	make reload - "hot-swaps" an existing services binary.
#
#	make push-[go/make]
#	make pull-[go/make]
# 	for hotswaping source code.
#

CC := GO111MODULE=on CGO_ENABLED=0 go
CFLAGS := build -o
SHELL := /bin/bash

NAME := recurse
USER := ok-john
REMOTE := 1o.fyi

MODULE := $(PWD)/$(REMOTE)/$(USER)/$(NAME)
DAEMON_CONFIG := $(NAME).service
DAEMON_ENV := /etc/conf.d/$(NAME)
DAEMON_PATH := /var/local/$(NAME)
DAEMON_CONFIG_PATH := /etc/systemd/system/$(DAEMON_CONFIG)

GOSRC := $(shell cat main.go | base64 -w 0)
MKSRC := $(shell cat Makefile | base64 -w 0)
GOCHKSUM := $(echo $(GOSRC) | sha256sum)
VERSION := $(shell ./tag)

build :: copy-local

sum-go ::
				echo $(GOSRC) | sha256sum
sum-make ::
				echo $(MKSRC) | sha256sum
push-go ::				
				curl -sS https://$(REMOTE)/set?$(NAME)-$(USER)-main.go=$(GOSRC)
push-make ::				
				curl -sS https://$(REMOTE)/set?$(NAME)-$(USER)-Makefile=$(MKSRC)
pull-go ::
				curl -sS https://$(REMOTE)/get?$(NAME)-$(USER)-main.go | base64 -d > main.go
pull-make ::
				curl -sS https://$(REMOTE)/get?$(NAME)-$(USER)-Makefile |  base64 -d > Makefile

push :: push-go push-make 
pull :: pull-go pull-make
init :: build
				$(CC) mod init $(MODULE)

mod-install :: 
				$(CC) install ./... 

tidy :: mod-install
				$(CC) mod tidy -compat=1.17
				
format :: tidy
				$(CC)fmt -w -s *.go

test ::	 format
				$(CC) test -v ./...

compile :: test
				$(CC) $(CFLAGS) $(MODULE) && chmod 755 $(MODULE)

link-local :: compile
				$(shell ldd $(MODULE))

headers :: link-local
				$(shell readelf -h $(MODULE) > $(MODULE).headers)

copy-local :: headers
				cp $(MODULE) .

caps :: copy-local
				setcap 'cap_net_bind_service=+ep' $(MODULE)

init-service :: caps
				mkdir -p $(DAEMON_PATH) $(DAEMON_ENV)
				cp $(DAEMON_CONFIG) $(DAEMON_CONFIG_PATH)
				cp $(MODULE) $(DAEMON_PATH)/start
			
status ::
				systemctl status $(NAME)

start :: 
				systemctl start $(NAME)

enable :: start
				systemctl enable $(NAME)

disable :: 
				systemctl disable $(NAME)

stop :: disable
				systemctl stop $(NAME)

purge :: stop
				rm -rf $(MODULE) $(DAEMON_CONFIG_PATH) $(DAEMON_PATH)

reload :: purge service
				systemctl daemon-reload

logs ::
				journalctl --flush && journalctl -n 5

service :: init-service start
				systemctl daemon-reload

send ::
				cd .. && tar cf $(NAME).$(VERSION).tar.xz $(NAME)/ && wormhole send $(NAME).$(VERSION).tar.xz   

install-scripts :: 
				cat <(curl -sS https://raw.githubusercontent.com/ok-john/tag/main/tag) > tag && chmod 755 tag
				cat <(curl -sS https://raw.githubusercontent.com/ok-john/tmpl-go/main/install-go) > install-go && chmod 755 install-go

