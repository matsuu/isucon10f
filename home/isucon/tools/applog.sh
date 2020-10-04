#!/bin/sh

ssh isu3.t.isucon.dev sudo journalctl -u xsuportal-web-golang -u xsuportal-api-golang -f
