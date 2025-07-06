package model

import (
	"encoding/json"
	"fmt"
	"os"
)

type Event struct {
	Device  string `json:"device"`
	Channel int    `json:"channel"`
	Key     int    `json:"key"`
	Value   int    `json:"value"`
}

type Command struct {
	Event     Event  `json:"event"`
	Alias     string `json:"alias"`
	Trigger   string `json:"trigger"`
	Command   string `json:"command"`
	TimeoutMs int    `json:"timeout_ms"`
}

type Commands struct {
	Commands []Command `json:"commands"`
	// ugly solution
	HighlightId int `json:"highlightId,omitempty"`
}

type Config struct {
	FilePath string
	Data     Commands
}

func NewConfig(FilePath string) (*Config, error) {
	file, err := os.Open(FilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{FilePath: FilePath}, nil
		}
		return nil, err
	}
	defer file.Close()

	var config Config
	config.FilePath = FilePath
	err = json.NewDecoder(file).Decode(&config.Data)
	return &config, err
}

func (c *Config) Print() {
	jsonBytes, err := json.MarshalIndent(c.Data, "", "  ")
	if err != nil {
		return
	}
	fmt.Println(string(jsonBytes))
}

func (c *Config) AddCommand(command Command) {
	c.Data.Commands = append(c.Data.Commands, command)
}

func (c *Config) DeleteCommand(commandIndex int) {
	if commandIndex >= 0 && commandIndex < len(c.Data.Commands) {
		c.Data.Commands = append(c.Data.Commands[:commandIndex], c.Data.Commands[commandIndex+1:]...)
	}
}

func (c *Config) ClearConfig() {
	c.Data.Commands = []Command{}
}

func (c *Config) GetEventId(in Event) (int, error) {
	for i, command := range c.Data.Commands {
		if command.Event.Device == in.Device &&
			command.Event.Channel == in.Channel &&
			command.Event.Key == in.Key {
			return i, nil
		}
	}
	return -1, fmt.Errorf("event %v not found", in)
}

func (c *Config) GetCommandWithId(in Event) (Command, int, error) {
	id, err := c.GetEventId(in)
	if err != nil {
		return Command{}, id, err
	}

	command := c.Data.Commands[id]
	if command.Trigger == "OnChange" ||
		(in.Value == 0 && command.Trigger == "OnRelease") ||
		(in.Value != 0 && command.Trigger == "OnPress") {
		return command, id, nil
	}

	return Command{}, -1, fmt.Errorf("trigger %v doesn't match", in)
}

func (c *Config) Save() error {
	jsonBytes, err := json.MarshalIndent(c.Data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(c.FilePath, jsonBytes, 0644)
}
