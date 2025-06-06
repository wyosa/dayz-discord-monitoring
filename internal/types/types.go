package types

type Server struct {
	Ip        string `yaml:"ip"`
	Port      int    `yaml:"port"`
	QueryPort int    `yaml:"query_port"`
}

type Emojis struct {
	Human string `yaml:"human"`
	Day   string `yaml:"day"`
	Night string `yaml:"night"`
}
