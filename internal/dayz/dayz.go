package dayz

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"time"
)

type Info struct {
	Server  ServerInfo
	Rules   ServerRules
	Players ServerPlayers
}

func GetServerInfo(ip string, port int) (Info, error) {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return Info{}, fmt.Errorf("Invalid IP address: %s", ip)
	}

	addr := &net.UDPAddr{
		IP:   parsedIP,
		Port: port,
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return Info{}, err
	}
	defer conn.Close()

	if err := conn.SetDeadline(time.Now().Add(4 * time.Second)); err != nil {
		return Info{}, err
	}

	server, err1 := getInfo(conn)
	rules, err2 := getRules(conn)
	players, err3 := getPlayers(conn)
	err = errors.Join(err1, err2, err3)
	if err != nil {
		return Info{}, err
	}

	fmt.Println(Info{
		Server:  server,
		Rules:   rules,
		Players: players,
	})

	return Info{
		Server:  server,
		Rules:   rules,
		Players: players,
	}, nil
}

type ServerInfo struct {
	Protocol    byte
	Name        string
	Map         string
	Folder      string
	Game        string
	ID          uint16
	Players     byte
	MaxPlayers  byte
	Bots        byte
	ServerType  byte
	Environment byte
	Visibility  byte
	VAC         byte
	Version     string
	EDF         byte
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

	data := bytes.NewBuffer(buffer[:n])
	info := ServerInfo{}

	// пропускаем первые 4 байта, потому что там лежит header
	for i := 0; i < 4; i++ {
		_, err := data.ReadByte()
		if err != nil {
			return ServerInfo{}, err
		}
	}

	binary.Read(data, binary.LittleEndian, &info.Protocol)
	info.Name, _ = readNullTerminatedString(data)
	info.Map, _ = readNullTerminatedString(data)
	info.Folder, _ = readNullTerminatedString(data)
	info.Game, _ = readNullTerminatedString(data)
	binary.Read(data, binary.LittleEndian, &info.ID)
	binary.Read(data, binary.LittleEndian, &info.Players)
	binary.Read(data, binary.LittleEndian, &info.MaxPlayers)
	binary.Read(data, binary.LittleEndian, &info.Bots)
	binary.Read(data, binary.LittleEndian, &info.ServerType)
	binary.Read(data, binary.LittleEndian, &info.Environment)
	binary.Read(data, binary.LittleEndian, &info.Visibility)
	binary.Read(data, binary.LittleEndian, &info.VAC)
	info.Version, _ = readNullTerminatedString(data)
	binary.Read(data, binary.LittleEndian, &info.EDF)

	return info, nil
}

type ServerRules struct{}

func getRules(conn *net.UDPConn) (ServerRules, error) {
	A2S_RULES_REQUEST := []byte{
		0xFF, 0xFF, 0xFF, 0xFF,
		'V',
		'S', 'o', 'u', 'r', 'c', 'e', ' ', 'E', 'n', 'g', 'i', 'n', 'e', ' ', 'Q', 'u', 'e', 'r', 'y',
		0x00,
	}

	_ = conn
	_ = A2S_RULES_REQUEST

	return ServerRules{}, nil
}

type ServerPlayers struct{}

func getPlayers(conn *net.UDPConn) (ServerPlayers, error) {
	A2S_PLAYER_REQUEST := []byte{
		0xFF, 0xFF, 0xFF, 0xFF,
		'U',
		'S', 'o', 'u', 'r', 'c', 'e', ' ', 'E', 'n', 'g', 'i', 'n', 'e', ' ', 'Q', 'u', 'e', 'r', 'y',
		0x00,
	}

	_ = conn
	_ = A2S_PLAYER_REQUEST

	return ServerPlayers{}, nil
}

func readNullTerminatedString(buf *bytes.Buffer) (string, error) {
	var result []byte
	for {
		b, err := buf.ReadByte()
		if err != nil {
			return "", err
		}
		if b == 0x00 {
			break
		}
		result = append(result, b)
	}
	return string(result), nil
}
