#!/bin/bash

go mod download

sudo cp .certs/charles-ssl.pem /usr/local/share/ca-certificates/charles-ssl-proxying-certificate.crt
sudo update-ca-certificates
