CC := GO111MODULE=on CGO_ENABLED=0 go
CFLAGS := build -o
SHELL := /bin/bash

# Provided values, Should be moved to a external, portable config.
REFLECT_PATH := /var/local
NAME := service-template
USER := ok-john
REMOTE := github.com

# Constructed values, should not be moved to a config.
MODULE := $(REMOTE)/$(USER)/$(NAME)
DAEMON_CONFIG := $(NAME).service
DAEMON_PATH := $(REFLECT_PATH)/$(NAME)
DAEMON_CONFIG_PATH := /etc/systemd/system/$(DAEMON_CONFIG)

# Building locally just use.
#		
# 		$ make
#		$ ./reflect
# 
# Linking as a service, will require sudo on most systems.
#
#		$ make service
#		
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


# These commands are used for creating/reloading a systemd service, you can build reflect
# and not use it in the form of a service.

init-service :: 
				mkdir -p $(DAEMON_PATH)

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

# Helper scripts - not neccessary at all.
install-scripts :: 
				cat <(curl -sS https://raw.githubusercontent.com/ok-john/tag/main/tag) > tag && chmod 755 tag
				cat <(curl -sS https://raw.githubusercontent.com/ok-john/tmpl-go/main/install-go) > install-go && chmod 755 install-go

# This will trace the execution of a local binary and parse an abstract sytanx tree of system-calls made by said-binary				
trace :: 
				trace-cmd record -p function_graph -F ./$(MODULE)
				trace-cmd report | sed 's/.* | //g' > $(MODULE).ttree
				trace-cmd record -p function_graph -e syscalls -F ./$(MODULE)
				trace-cmd report | sed 's/.* | //g' > $(MODULE)-syscalls.ttree
				trace-cmd record -p function_graph -g __x64_sys_read ./$(MODULE)
				trace-cmd report | sed 's/.* | //g' > $(MODULE)-sysreads.ttree

