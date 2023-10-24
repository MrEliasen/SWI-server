#!/bin/bash

mkdir -p ./certs
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
    -keyout certs/localhost.key -out certs/localhost.crt \
    -subj "/C=GB/ST=London/L=London/O=Dev Environment/OU=DevEnv/CN=localhost/emailAddress=dev.environment@localhost.tld"

