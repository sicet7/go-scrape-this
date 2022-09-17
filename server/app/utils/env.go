package utils

import (
	"errors"
	"os"
	"strconv"
	"strings"
)

func RequireIntEnv(name string) (int, error) {
	value, ok := os.LookupEnv(name)
	if !ok || len(strings.TrimSpace(value)) == 0 {
		return 0, errors.New("missing required environment variable: \"" + name + "\"")
	}
	intVal, err := strconv.Atoi(value)
	if err != nil {
		return 0, errors.New("invalid environment variable: \"" + name + "\" is not a integer value")
	}
	return intVal, nil
}

func RequireStringEnv(name string) (string, error) {
	value, ok := os.LookupEnv(name)
	if !ok || len(strings.TrimSpace(value)) == 0 {
		return "", errors.New("missing required environment variable: \"" + name + "\"")
	}
	return strings.TrimSpace(value), nil
}

func ReadIntEnv(name string, defaultValue int) int {
	value, ok := os.LookupEnv(name)
	if !ok || len(strings.TrimSpace(value)) == 0 {
		return defaultValue
	}
	intVal, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intVal
}

func ReadStringEnv(name string, defaultValue string) string {
	value, ok := os.LookupEnv(name)
	if !ok || len(strings.TrimSpace(value)) == 0 {
		return defaultValue
	}
	return strings.TrimSpace(value)
}

func parseBoolValue(value string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "true":
	case "1":
	case "yes":
	case "y":
		return true, nil
	case "false":
	case "0":
	case "no":
	case "n":
		return false, nil
	}
	return false, errors.New("failed to parse environment variable value")
}

func RequireBoolEnv(name string) (bool, error) {
	value, err := RequireStringEnv(name)
	if err != nil {
		return false, err
	}
	output, err := parseBoolValue(value)
	if err != nil {
		return false, errors.New("failed to parse environment variable: \"" + name + "\"")
	}
	return output, nil
}

func ReadBoolEnv(name string, defaultValue bool) bool {
	var value string
	if defaultValue {
		value = ReadStringEnv(name, "true")
	} else {
		value = ReadStringEnv(name, "false")
	}
	output, err := parseBoolValue(value)
	if err != nil {
		return defaultValue
	}
	return output
}
