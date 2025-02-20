#!/bin/sh

DELOS_PORT=42586
KVSTORE_PORT=42585

./kvstore --tcp-serv-port $KVSTORE_PORT --tcp-dial-port $DELOS_PORT &
./delos --tcp-serv-port $DELOS_PORT &
