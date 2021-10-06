## LogWatcher

[![Build](https://github.com/CondensedTea/LogWatcher/actions/workflows/ci.yaml/badge.svg?branch=main)](https://github.com/CondensedTea/LogWatcher/actions/workflows/build.yaml)
[![codecov](https://codecov.io/gh/CondensedTea/LogWatcher/branch/main/graph/badge.svg?token=X1VK4MWV16)](https://codecov.io/gh/CondensedTea/LogWatcher)

Service for centralised log collection from tf2 servers.

Made for tf2pickup.org project.

Features:

* Collect logs from multiple servers
* Upload logs to https://logs.tf.
* Save basic players stats to mongodb.

#### How to use:

1. Set up your TF2 server to send logs by UDP:

```
logaddress_add <logwatcher-IP-address>:27100
```

3. Create your config with `config.template.yaml`

4. Build Docker image and run server on 27000/udp:

```bash
make build run
```
