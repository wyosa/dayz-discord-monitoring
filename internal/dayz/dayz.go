package dayz

import (
	"bytes"
	"encoding/binary"
	"errors"
	"net"
	"strings"
	"time"
)

type ServerInfo struct {
	Protocol    byte
	Name        string
	Map         string
	Folder      string
	Game        string
	ID          uint16
	Players     byte
	MaxPlayers  byte
	Time        string
	Queue       string
	Bots        byte
	ServerType  byte
	Environment byte
	Visibility  byte
	VAC         byte
	Version     string
	EDF         byte
	Keywords    string
}

func GetServerInfo(ip string, port int) func() (int, ServerInfo, error) {
	timeoutsCounter := 0

	return func() (int, ServerInfo, error) {
		parsedIP := net.ParseIP(ip)
		if parsedIP == nil {
			timeoutsCounter++
			return timeoutsCounter, ServerInfo{}, errors.New("invalid IP address")
		}

		addr := &net.UDPAddr{
			IP:   parsedIP,
			Port: port,
		}

		conn, err := net.DialUDP("udp", nil, addr)
		if err != nil {
			timeoutsCounter++
			return timeoutsCounter, ServerInfo{}, err
		}
		defer conn.Close()

		if err := conn.SetDeadline(time.Now().Add(4 * time.Second)); err != nil {
			timeoutsCounter++
			return timeoutsCounter, ServerInfo{}, err
		}

		info, err := getInfo(conn)
		if err != nil {
			timeoutsCounter++
			return timeoutsCounter, ServerInfo{}, err
		}

		timeoutsCounter = 0
		return timeoutsCounter, info, nil
	}
}

func getInfo(conn *net.UDPConn) (ServerInfo, error) {
	A2S_INFO_REQUEST := []byte{
		0xFF, 0xFF, 0xFF, 0xFF,
		'T',
		'S', 'o', 'u', 'r', 'c', 'e', ' ', 'E', 'n', 'g', 'i', 'n', 'e', ' ', 'Q', 'u', 'e', 'r', 'y',
		0x00,
	}

	conn.Write(A2S_INFO_REQUEST)

	buffer := make([]byte, 1400)
	n, _, err := conn.ReadFromUDP(buffer)
	if err != nil {
		return ServerInfo{}, err
	}

	if n < 5 {
		return ServerInfo{}, errors.New("response too short")
	}

	buf := bytes.NewBuffer(buffer[:n])
	info := ServerInfo{}

	// Skip 4-byte header
	buf.Next(4)

	// Read protocol version
	if err := binary.Read(buf, binary.LittleEndian, &info.Protocol); err != nil {
		return info, err
	}

	if info.Name, err = buf.ReadString(0x00); err != nil {
		return info, err
	}
	if info.Map, err = buf.ReadString(0x00); err != nil {
		return info, err
	}
	if info.Folder, err = buf.ReadString(0x00); err != nil {
		return info, err
	}
	if info.Game, err = buf.ReadString(0x00); err != nil {
		return info, err
	}

	if err := binary.Read(buf, binary.LittleEndian, &info.ID); err != nil {
		return info, err
	}
	if err := binary.Read(buf, binary.LittleEndian, &info.Players); err != nil {
		return info, err
	}
	if err := binary.Read(buf, binary.LittleEndian, &info.MaxPlayers); err != nil {
		return info, err
	}
	if err := binary.Read(buf, binary.LittleEndian, &info.Bots); err != nil {
		return info, err
	}
	if err := binary.Read(buf, binary.LittleEndian, &info.ServerType); err != nil {
		return info, err
	}
	if err := binary.Read(buf, binary.LittleEndian, &info.Environment); err != nil {
		return info, err
	}
	if err := binary.Read(buf, binary.LittleEndian, &info.Visibility); err != nil {
		return info, err
	}
	if err := binary.Read(buf, binary.LittleEndian, &info.VAC); err != nil {
		return info, err
	}

	if info.Version, err = buf.ReadString(0x00); err != nil {
		return info, err
	}

	// Optional EDF
	if buf.Len() > 0 {
		if err := binary.Read(buf, binary.LittleEndian, &info.EDF); err != nil {
			return info, err
		}

		if info.EDF&0x80 != 0 {
			var port uint16
			if err := binary.Read(buf, binary.LittleEndian, &port); err != nil {
				return info, err
			}
		}
		if info.EDF&0x10 != 0 {
			var steamID uint64
			if err := binary.Read(buf, binary.LittleEndian, &steamID); err != nil {
				return info, err
			}
		}
		if info.EDF&0x40 != 0 {
			var tvPort uint16
			if err := binary.Read(buf, binary.LittleEndian, &tvPort); err != nil {
				return info, err
			}
			if _, err := buf.ReadString(0x00); err != nil {
				return info, err
			}
		}
		if info.EDF&0x20 != 0 {
			if info.Keywords, err = buf.ReadString(0x00); err != nil {
				return info, err
			}
		}
		if info.EDF&0x01 != 0 {
			var gameID uint64
			if err := binary.Read(buf, binary.LittleEndian, &gameID); err != nil {
				return info, err
			}
		}
	}

	if info.Keywords != "" {
		keywords := strings.Split(info.Keywords, ",")
		for _, keyword := range keywords {
			// Find queue
			if strings.Contains(keyword, "lqs") {
				info.Queue = strings.TrimPrefix(keyword, "lqs")
			}

			// Find time
			if strings.Contains(keyword, ":") {
				info.Time = keyword
			}
		}
	}

	return info, nil
}
