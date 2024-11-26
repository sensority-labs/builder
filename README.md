Watchman builder

# Installation
To work with private repos you need to allow them:

Add to your rc-file (.bashrc or .zshrc or similar):
```shell
export GOPRIVATE=github.com/sensority-labs/*
export PATH=$PATH:$HOME/go/bin
```
Then install binary:
```bash
go install github.com/sensority-labs/builder/cmd/bot-builder@latest
```

# Usage
Env params:
- `NETWORK_NAME` - docker network name to connect to. Default is `sensority-labs`
- `NATS_URL` - nats url to connect to. Default is `nats://nats:4222`
- `PORT` - port to listen on. Default is `5005`

With defaults:
```bash
watchman-builder
```

With custom settings:
```bash
NETWORK_NAME=watchman-builder-network NATS_URL=nats://localhost:4222 bot-builder
```