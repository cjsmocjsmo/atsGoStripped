#!/bin/sh

openssl req  -new  -newkey rsa:2048  -nodes  -keyout alphatree.key  -out alphatree.csr && \
openssl  x509  -req  -days 365  -in alphatree.csr  -signkey alphatree.key  -out alphatree.crt && \
docker-compose up -d --build

# apt-get install -y snapd && \
# snap install core && \ 
# snap refresh core && \
# snap install --classic certbot && \
# certbot certonly --standalone