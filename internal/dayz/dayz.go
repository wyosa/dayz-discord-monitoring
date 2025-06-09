package dayz

import (
	"bytes"
	"encoding/binary"
	"errors"
	"net"
	"strings"
	"time"
)

// Server represents a DayZ server connection configuration.
type Server struct {
	IP        string // Server IP address
	QueryPort int    // Query port for A2S_INFO requests
}

// ServerInfo contains all information returned by A2S_INFO query.
// Based on Source Engine Query protocol specification.
type ServerInfo struct {
	Protocol    byte   // Protocol version used by the server
	Name        string // Server name
	Map         string // Current map name
	Folder      string // Game directory name
	Game        string // Game name
	ID          uint16 // Application ID of the game
	Players     byte   // Current number of players
	MaxPlayers  byte   // Maximum number of players
	Time        string // Current in-game time (parsed from keywords)
	Queue       string // Queue length (parsed from keywords)
	Bots        byte   // Number of bots on the server
	ServerType  byte   // Server type (dedicated/listen/proxy)
	Environment byte   // Server environment (Linux/Windows/Mac)
	Visibility  byte   // Server visibility (public/private)
	VAC         byte   // VAC status (secured/unsecured)
	Version     string // Server version
	EDF         byte   // Extra Data Flag - indicates which optional fields are present
	Keywords    string // Server keywords containing custom data
	Port        uint16 // Server port number
	SteamID     uint64 // Server's Steam ID
	TVPort      uint16 // Source TV port
	GameID      uint64 // Game ID
}

const (
	headerByteSize int = 4
	zeroByte byte = 0x00
	timeoutTime = 3
	optimalBufferSize int = 1400
)

// GetServerInfo queries the DayZ server for current status information.
// Uses the A2S_INFO query to retrieve server details.
func (server Server) GetServerInfo() (ServerInfo, error) {
	parsedIP := net.ParseIP(server.IP)
	if parsedIP == nil {
		return ServerInfo{}, errors.New("invalid IP address")
	}

	addr := &net.UDPAddr{
		IP:   parsedIP,
		Port: server.QueryPort,
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return ServerInfo{}, err
	}
	defer conn.Close()

	if err = conn.SetDeadline(time.Now().Add(timeoutTime * time.Second)); err != nil {
		return ServerInfo{}, err
	}

	a2sInfoRequest := []byte{
		0xFF, 0xFF, 0xFF, 0xFF,
		'T',
		'S', 'o', 'u', 'r', 'c', 'e', ' ', 'E', 'n', 'g', 'i', 'n', 'e', ' ', 'Q', 'u', 'e', 'r', 'y',
		0x00,
	}

	_, err = conn.Write(a2sInfoRequest)
	if err != nil {
		return ServerInfo{}, err
	}

	buffer := make([]byte, optimalBufferSize)
	n, _, err := conn.ReadFromUDP(buffer)
	if err != nil {
		return ServerInfo{}, err
	}

	if n <= headerByteSize {
		return ServerInfo{}, errors.New("response too short")
	}

	buf := bytes.NewBuffer(buffer[:n])
	info := ServerInfo{}

	// Skip the 4-byte response header
	buf.Next(headerByteSize)

	// Parse required fields
	if err = binary.Read(buf, binary.LittleEndian, &info.Protocol); err != nil {
		return info, err
	}
	if info.Name, err = buf.ReadString(zeroByte); err != nil {
		return info, err
	}
	if info.Map, err = buf.ReadString(zeroByte); err != nil {
		return info, err
	}
	if info.Folder, err = buf.ReadString(zeroByte); err != nil {
		return info, err
	}
	if info.Game, err = buf.ReadString(zeroByte); err != nil {
		return info, err
	}
	if err = binary.Read(buf, binary.LittleEndian, &info.ID); err != nil {
		return info, err
	}
	if err = binary.Read(buf, binary.LittleEndian, &info.Players); err != nil {
		return info, err
	}
	if err = binary.Read(buf, binary.LittleEndian, &info.MaxPlayers); err != nil {
		return info, err
	}
	if err = binary.Read(buf, binary.LittleEndian, &info.Bots); err != nil {
		return info, err
	}
	if err = binary.Read(buf, binary.LittleEndian, &info.ServerType); err != nil {
		return info, err
	}
	if err = binary.Read(buf, binary.LittleEndian, &info.Environment); err != nil {
		return info, err
	}
	if err = binary.Read(buf, binary.LittleEndian, &info.Visibility); err != nil {
		return info, err
	}
	if err = binary.Read(buf, binary.LittleEndian, &info.VAC); err != nil {
		return info, err
	}
	if info.Version, err = buf.ReadString(zeroByte); err != nil {
		return info, err
	}

	// Parse optional EDF
	if buf.Len() > 0 {
		if err = binary.Read(buf, binary.LittleEndian, &info.EDF); err != nil {
			return info, err
		}

		if info.EDF&0x80 != 0 {
			if err = binary.Read(buf, binary.LittleEndian, &info.Port); err != nil {
				return info, err
			}
		}

		if info.EDF&0x10 != 0 {
			if err = binary.Read(buf, binary.LittleEndian, &info.SteamID); err != nil {
				return info, err
			}
		}
		if info.EDF&0x40 != 0 {
			if err = binary.Read(buf, binary.LittleEndian, &info.TVPort); err != nil {
				return info, err
			}
			if _, err = buf.ReadString(zeroByte); err != nil {
				return info, err
			}
		}
		if info.EDF&0x20 != 0 {
			if info.Keywords, err = buf.ReadString(zeroByte); err != nil {
				return info, err
			}
		}
		if info.EDF&0x01 != 0 {
			if err = binary.Read(buf, binary.LittleEndian, &info.GameID); err != nil {
				return info, err
			}
		}
	}

	// Parse DayZ-specific information from keywords
	if info.Keywords != "" {
		keywords := strings.Split(info.Keywords, ",")

		for _, keyword := range keywords {
			// Extract queue length (format: "lqs<number>")
			if strings.Contains(keyword, "lqs") {
				info.Queue = strings.TrimPrefix(keyword, "lqs")
			}

			// Extract in-game time (format: "HH:MM")
			if strings.Contains(keyword, ":") {
				info.Time = keyword
			}
		}
	}

	return info, nil
}
