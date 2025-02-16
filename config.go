package main

type ServerConfig struct {
	Prefix        string `json:"prefix"`
	TargetChannel string `json:"target_channel"`
	GuildID       string `json:"guildID"`
}

func (sc *ServerConfig) SetTargetChannel(id string) {
	sc.TargetChannel = id
}

type DSConfig struct {
	BotToken string         `json:"token"`
	Servers  []ServerConfig `json:"servers"`
}

func (d *DSConfig) GetServer(id string) *ServerConfig {
	for _, s := range d.Servers {
		if s.GuildID == id {
			return &s
		}
	}
	// If we get to this point I am assuming that we didn't find a suitable server
	// So now we are going to create a new one, add it the main config, and return a reference
	newServer := ServerConfig{
		Prefix:  "!",
		GuildID: id,
	}
	d.Servers = append(d.Servers, newServer)
	return d.GetServer(id)
}
