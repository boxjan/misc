package main

type Config struct {
	Addr string
}

var defaultConfig = Config{
	Addr: "[::]:8000",
}
