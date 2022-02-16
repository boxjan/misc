package main

type Config struct {
	Addr        string
	ExtraFormat map[string]string
}

var defaultConfig = Config{
	Addr: "[::]:8000",
	ExtraFormat: map[string]string{
		"RFC3339": "2006-01-02T15:04:05Z07:00",
	},
}
