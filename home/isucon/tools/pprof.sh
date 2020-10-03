#!/bin/sh

go tool pprof -no_browser -http :8888 /home/isucon/isuumo/webapp/go/isuumo 127.0.0.1:1323/debug/pprof/profile?seconds=30

