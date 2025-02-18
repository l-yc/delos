#!/bin/sh

echo "val of a: "
curl -L http://127.0.0.1:1337/a
echo

curl -L http://127.0.0.1:1337/a -XPUT -d 2

echo "val of a (expect 2): "
curl -L http://127.0.0.1:1337/a
echo
