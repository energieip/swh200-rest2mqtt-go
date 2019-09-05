#!/bin/bash

export PATH=$PATH:/usr/local/bin

op="${1:-op}"
mac="${2:-mac}"
ip="${3:-ip}"
hostname="${4}"

oid=${mac:0:8}
oid=$(echo ${oid^^})

if [ "$oid" == "F2:23:D0" ]; then
  echo "$(date) EIP device skip it $mac" >> /tmp/log.txt
  exit 0
fi

echo "$(date) Get $1 IP: $ip MAC: $mac" >> /tmp/log.txt

$(new-device -c /etc/energieip-swh200-rest2mqtt/config.json -i "$ip" -m "$mac")
