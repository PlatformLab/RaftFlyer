package utils

import (
    "os"
    "bufio"
    "fmt"
    "raft"
)

type Config struct {
    Servers []raft.ServerAddress
}

func ReadConfig(path string) (*Config, error) {
    file, err := os.Open(path)
    if err != nil {
        return nil, fmt.Errorf("Cannot open path: %s", path)
    }
    defer file.Close()

    var lines []raft.ServerAddress
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        lines = append(lines, raft.ServerAddress(scanner.Text()))
    }

    c := &Config {
        Servers: lines,
    }
    return c, nil
}
