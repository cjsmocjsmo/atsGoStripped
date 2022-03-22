#!/bin/sh

curl -fsSL https://download.docker.com/linux/debian/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg && \
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/debian \
  $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null && \

apt-get update && \
apt-get dist-upgrade -y && \
apt-get auto-clean -y && \
apt-get auto-remove -y && \



apt-get install -y docker-ce docker-ce-cli containerd.io snapd nano && \
snap install core && \ 
snap refresh core && \
snap install --classic certbot && \
certbot certonly --standalone
# openssl req  -new  -newkey rsa:2048  -nodes  -keyout alphatree.key  -out alphatree.csr && \
# openssl  x509  -req  -days 365  -in alphatree.csr  -signkey alphatree.key  -out alphatree.crt && \
apt-get install -y docker-compose
# docker-compose up -d --build
