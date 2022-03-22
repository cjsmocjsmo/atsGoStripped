#!/bin/sh

apt-get update && \
apt-get dist-upgrade -y && \
apt-get auto-clean -y && \
apt-get auto-remove -y && \
apt-get install -y docker-compose openssl ca-certificates curl gnupg lsb-release && \
curl -fsSL https://download.docker.com/linux/debian/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg && \
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/debian \
  $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null && \


apt-get install docker-ce docker-ce-cli containerd.io -y && \
openssl req  -new  -newkey rsa:2048  -nodes  -keyout alphatree.key  -out alphatree.csr && \
openssl  x509  -req  -days 365  -in alphatree.csr  -signkey alphatree.key  -out alphatree.crt && \
docker-compose up -d --build
# snapd && \
# snap install core && \ 
# snap refresh core && \
# snap install --classic certbot && \
# certbot certonly --standalone