Watchman builder

# Installation
```bash
go install github.com/sensority-labs/builder/cmd/watchman-builder@latest
```

# Usage
Env params:
- `NETWORK_NAME` - docker network name to connect to. Default is `sensority-labs`
- `NATS_URL` - nats url to connect to. Default is `nats://nats:4222`

With defaults:
```bash
watchman-builder
```

With custom settings:
```bash
NETWORK_NAME=watchman-builder-network NATS_URL=nats://localhost:4222 watchman-builder
```