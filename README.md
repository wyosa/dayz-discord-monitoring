<div align="center">
<img src="./.readme/assets/profile.png" alt="Profile view"/>
<br />
<br />
<img src="./.readme/assets/sidebar.png" alt="Sidebar view"/>
<br />
<br />
</div>

A lightweight Go tool that launches multiple Discord bots to monitor DayZ servers in real-time.

## Requirements

- **[Golang](https://go.dev/doc/install)** (version 1.24.4 or later)
- A **DayZ** server with accessible IP and query port
- **Discord bot token** from [Discord Developer Portal](https://discord.com/developers/applications) with Administrator permissions (or at least change nickname and change presence permissions)

## Quick Start

#### 1. Clone and navigate to the project

```bash
git clone https://github.com/wyosa/dayz-discord-monitoring.git

cd dayz-discord-monitoring
```

#### 2. Configure your bots

```bash
cp config/config.example.yaml config/config.yaml
# Edit config/config.yaml with your bot tokens and server details
```

#### 3. Build

```bash
go build -o bots ./cmd/bots

```

#### 4. Run

```bash
# Windows
bots.exe -config="config/config.yaml"

# Linux/MacOS
./bots -config="config/config.yaml"
```

## Configuration Options

| Field                      | Description                         | Example                                                                |
| -------------------------- | ----------------------------------- | ---------------------------------------------------------------------- |
| `offline`                  | Status text when server is offline  | `"Server offline"`                                                     |
| `emojis.human`             | Emoji for player count              | `"👤"`                                                                 |
| `emojis.day`               | Emoji for daytime (06:00-19:59)     | `"☀️"`                                                                 |
| `emojis.night`             | Emoji for nighttime (20:00-05:59)   | `"🌕"`                                                                 |
| `bots[].name`              | Bot display name (max 32 chars)     | `"My Server"`                                                          |
| `bots[].discord_token`     | Your Discord bot token              | Get from [Discord Portal](https://discord.com/developers/applications) |
| `bots[].update_interval`   | Update frequency in seconds (min 5) | `10`                                                                   |
| `bots[].server.ip`         | DayZ server IP address              | `"123.456.789.0"`                                                      |
| `bots[].server.port`       | DayZ server port                    | `2302`                                                                 |
| `bots[].server.query_port` | DayZ server query port              | `27016`                                                                |
