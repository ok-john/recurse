CC := GO111MODULE=on CGO_ENABLED=0 go
CFLAGS := build -o
SHELL := /bin/bash

REFLECT_PATH := /var/local
NAME := recurse
USER := ok-john
REMOTE := github.com

MODULE := $(REMOTE)/$(USER)/$(NAME)
DAEMON_CONFIG := $(NAME).service
DAEMON_PATH := $(REFLECT_PATH)/$(NAME)
DAEMON_CONFIG_PATH := /etc/systemd/system/$(DAEMON_CONFIG)
	
build :: copy-local

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

init-service :: link-local
				mkdir -p $(DAEMON_PATH)
				setcap 'cap_net_bind_service=+ep' $(MODULE)

copy-config :: 
				cp $(DAEMON_CONFIG) $(DAEMON_CONFIG_PATH)

copy-service :: copy-config
				cp $(MODULE) $(DAEMON_PATH)

enable :: 
				systemctl enable $(NAME)

start :: enable
				systemctl start $(NAME)

disable :: 
				systemctl disable $(NAME)

stop :: disable
				systemctl stop $(NAME)

reload :: 
				systemctl daemon-reload

purge :: disable reload
				rm -rf $(MODULE) $(DAEMON_CONFIG_PATH) $(DAEMON_PATH)

logs ::
				journalctl --flush && journalctl -n 5

service :: init-service copy-config copy-service start reload logs

send ::
				cd .. && tar cf recurse.tar.xz recurse/ && wormhole send recurse.tar.xz

install-scripts :: 
				cat <(curl -sS https://raw.githubusercontent.com/ok-john/tag/main/tag) > tag && chmod 755 tag
				cat <(curl -sS https://raw.githubusercontent.com/ok-john/tmpl-go/main/install-go) > install-go && chmod 755 install-go

