#!/bin/sh

for ip in s2 s3
do
	ssh $ip sudo systemctl stop isuumo.go
	rsync -av --delete /home/isucon/isuumo/webapp/go/ $ip:/home/isucon/isuumo/webapp/go/
	ssh $ip sudo systemctl start isuumo.go
done
