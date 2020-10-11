#!/bin/sh

go tool pprof -no_browser -http :8888 /home/isucon/webapp/golang/bin/xsuportal isu2.t.isucon.dev:9292/debug/pprof/profile?seconds=30

