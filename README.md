Switch REST2MQTT: Convert REST devices to MQTTs drivers
==========================================================

Build Requirement: 
* golang-go > 1.9
* glide
* devscripts
* make

Run dependancies:
* mosquitto

To compile it:
* GOPATH needs to be configured, for example:
```
    export GOPATH=$HOME/go
```

* Install go dependancies:
```
    make prepare
```

* To clean build tree:
```
    make clean
```

* Multi-target build:
```
    make all
```

* To build x86 target:
```
    make bin/energieip-swh200-rest2mqtt-amd64
```

* To build armhf target:
```
    make bin/energieip-swh200-rest2mqtt-armhf
```
* To create debian archive for arm:
```
    make deb-armhf
```

* To install debian archive on the target:
```
    scp build/*.deb <login>@<ip>:~/
    ssh <login>@<ip>
    sudo dpkg -i *.deb
```

For development:
* recommanded logger: *rlog*
* For dependency: use *common-components-go* library
