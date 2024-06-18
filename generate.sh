#!/bin/bash

echo "Create QR codes by curl calling QR code API"
echo
# Check cURL command if available (required), abort if does not exists
type curl >/dev/null 2>&1 || { echo >&2 "curl required but not installed. Aborting."; exit 1; }
echo


while read -r i;do 

    URL=$i
    NAME="${i#http*://}"
    RESPONSE="curl -X POST --form "size=256" --form "url=${URL}" --output data/${NAME}.png http://localhost:8080/generate"
    echo "URL: $URL"
    echo "NAME: $NAME";
    echo "RESPONSE: $RESPONSE";
    `$RESPONSE`
done<example.txt
