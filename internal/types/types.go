package types

// Server represents the configuration for connecting to a DayZ game server.
type Server struct {
	IP        string `yaml:"ip"`         // Server IP address (e.g., "192.168.1.100")
	Port      int    `yaml:"port"`       // Game connection port (usually 2302 for DayZ)
	QueryPort int    `yaml:"query_port"` // A2S_INFO query port (usually 27016 for Source Engine)
}

// Emojis defines the emoji characters used in Discord bot status messages.
type Emojis struct {
	Human string `yaml:"human"` // Emoji for player count display (e.g., "👤", "🧑")
	Day   string `yaml:"day"`   // Emoji for daytime hours 06:00-19:59 (e.g., "🌞", "☀️")
	Night string `yaml:"night"` // Emoji for nighttime hours 20:00-05:59 (e.g., "🌙", "🌛")
}
