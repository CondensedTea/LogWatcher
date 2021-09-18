### LogWatcher
UDP server for log collection and upload to https://logs.tf

#### How to use: 
1. Setup your TF2 server to send logs to LogWatcher: 
```
logaddress_add <logwatcher-IP-address>:27100
```

2. Build Docker image:
```bash
make build
```
3. Run container with latest image:
```
make run
```