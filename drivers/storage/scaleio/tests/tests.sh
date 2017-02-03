#!/usr/bin/env bash

# This script runs tests

docker volume create --driver rexray --name myvoltest --opt size=16 >> /tmp/output.txt
docker volume rm myvoltest >> /tmp/output.txt
