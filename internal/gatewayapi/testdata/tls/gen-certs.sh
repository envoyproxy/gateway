#!/bin/bash

# The following commands were used to generate test key/cert pairs
# using openssl (LibreSSL 3.3.6)

CERT_VALIDITY_DAYS=3650

# RSA

openssl req -x509 -nodes -days $CERT_VALIDITY_DAYS -newkey rsa:2048 -keyout rsa-pkcs8.key -out rsa-cert.pem -subj "/CN=foo.bar.com"
openssl rsa -in rsa-pkcs8.key -out rsa-pkcs1.key

# RSA with SAN extension

openssl req -x509 -nodes -days $CERT_VALIDITY_DAYS -newkey rsa:2048 -keyout rsa-pkcs8-san.key -out rsa-cert-san.pem -subj "/CN=Test Inc" -addext "subjectAltName = DNS:foo.bar.com"
openssl rsa -in rsa-pkcs8-san.key -out rsa-pkcs1-san.key

# RSA with wildcard SAN domain

openssl req -x509 -nodes -days $CERT_VALIDITY_DAYS -newkey rsa:2048 -keyout rsa-pkcs8-wildcard.key -out rsa-cert-wildcard.pem -subj "/CN=Test Inc" -addext "subjectAltName = DNS:*, DNS:*.example.com"
openssl rsa -in rsa-pkcs8-wildcard.key -out rsa-pkcs1-wildcard.key

# ECDSA-p256

openssl ecparam -name prime256v1 -genkey -noout -out ecdsa-p256.key
openssl req -new -x509 -days $CERT_VALIDITY_DAYS -key ecdsa-p256.key -out ecdsa-p256-cert.pem -subj "/CN=foo.bar.com"

# ECDSA-p384

openssl ecparam -name secp384r1 -genkey -noout -out ecdsa-p384.key
openssl req -new -x509 -days $CERT_VALIDITY_DAYS -key ecdsa-p384.key -out ecdsa-p384-cert.pem -subj "/CN=foo.bar.com"
