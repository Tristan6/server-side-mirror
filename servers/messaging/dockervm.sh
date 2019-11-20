#!/usr/bin/env bash

export TLSCERT=/etc/letsencrypt/live/api.sauravkharb.me/fullchain.pem
export TLSKEY=/etc/letsencrypt/live/api.sauravkharb.me/privkey.pem

sudo docker rm -f messaging
sudo docker rm -f messaging2
sudo docker rm -f mongodb

sudo docker pull jtanderson7/assignment2

sudo docker run -d \
--network=messagingNetwork \
--name mongodb \
-v ~/data:/data/db \
mongo

sudo docker run -d \
--restart=unless-stopped \
--network=messagingNetwork \
--name messaging \
-v /etc/letsencrypt:/etc/letsencrypt:ro \
-e TLSCERT=$TLSCERT \
-e TLSKEY=$TLSKEY \
-e ADDR=5001 \
jtanderson7/assignment2

sudo docker run -d \
--restart=unless-stopped \
--network=messagingNetwork \
--name messaging2 \
-v /etc/letsencrypt:/etc/letsencrypt:ro \
-e TLSCERT=$TLSCERT \
-e TLSKEY=$TLSKEY \
-e ADDR=5002 \
jtanderson7/assignment2

docker ps