#!/bin/bash

export PATH=$PATH:/usr/local/bin

# script to detect new dhcp lease

# this will be called by dnsmasq everytime a new device is connected
# with the following arguments
# $1 = add | old
# $2 = mac address
# $3 = ip address
# $4 = device name
op="${1:-op}"
mac="${2:-mac}"
ip="${3:-ip}"
hostname="${4}"

oid=${mac:0:8}
oid=$(echo ${oid^^})

if [ "$oid" == "F2:23:D0" ]; then
  exit 0
fi

#echo "$(date) Get $1 IP: $ip MAC: $mac" >> /tmp/log.txt

if  [[ $mac == *":"* ]]; then
  $(new-device -c /etc/energieip-swh200-rest2mqtt/config.json -i "$ip" -m "$mac")
else
  echo "$(date) Invalid mac found $mac" >> /tmp/log.txt
fi
