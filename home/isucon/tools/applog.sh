#!/bin/sh

ssh isu2.t.isucon.dev sudo journalctl -u xsuportal-web-golang -u xsuportal-api-golang -f
