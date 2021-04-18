#!/bin/zsh

go build -o bookings cmd/web/*.go && ./bookings -dbname=bookings -dbuser=limyoungjae -cache=false -production=false
