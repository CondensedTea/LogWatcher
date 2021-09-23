## LogWatcher
[![Build](https://github.com/CondensedTea/LogWatcher/actions/workflows/build.yaml/badge.svg?branch=main)](https://github.com/CondensedTea/LogWatcher/actions/workflows/build.yaml)


UDP server for centralised log collection and upload to https://logs.tf

#### How to use: 
1. Setup your TF2 server to send logs to LogWatcher: 
```
logaddress_add <logwatcher-IP-address>:27100
```

2. Build Docker image and run server on 27000/udp:
```bash
make
```
