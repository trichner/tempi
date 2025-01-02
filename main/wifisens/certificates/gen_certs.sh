#!/usr/bin/env bash

set -xe

touch index.txt serial
touch crlnumber
echo 01 > serial
echo 1000 > crlnumber

#---- CA
openssl genrsa  -out cakey.pem 2048
openssl req -new -x509 -days 3650  -config cacert.cnf  -key cakey.pem -out cacert.pem

# verify the rootCA certificate content and X.509 extensions
openssl x509 -noout -text -in cacert.pem

#---- create server

openssl genrsa -out server.key 2048

openssl genrsa -out server.key 2048
openssl req -new -key server.key -out server.csr -config server_cert.cnf
openssl req -noout -text -in server.csr

#---- ca sign
#openssl x509 -req -days 365 -in server.csr -CA cacert.pem -CAkey cakey.pem -CAcreateserial -out server.crt
openssl x509 -req -days 365 -in server.csr -CA cacert.pem -CAkey cakey.pem -CAcreateserial -out server.crt -extensions req_ext -extfile server_cert.cnf

cat server.crt cacert.pem > chain.crt