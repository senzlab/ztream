package main

import (
    "os"
)

type Config struct {
	switchName  string
    switchPort  string
    switchMode  string
}

var config = Config {
    switchName: getEnv("SWITCH_NAME", "senzswitch"),
    switchPort: getEnv("SWITCH_PORT", "9090"),
    switchMode: getEnv("SWITCH_MODE", "dev"),
}

func getEnv(key, fallback string) string {
    if value, ok := os.LookupEnv(key); ok {
        return value
    }

    return fallback
}
