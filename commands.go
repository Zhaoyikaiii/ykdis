package main

import (
	"fmt"
	"strconv"
)

func ping(args []string) []byte {
	if len(args) == 0 {
		return []byte("+PONG\r\n")
	} else if len(args) == 1 {
		return []byte("+" + args[0] + "\r\n")
	} else {
		return []byte("-ERR wrong number of arguments for 'ping' command\r\n")
	}
}

func echo(args []string) []byte {
	if len(args) != 1 {
		return []byte("-ERR wrong number of arguments for 'echo' command\r\n")
	}
	return []byte("$" + fmt.Sprintf("%d", len(args[0])) + "\r\n" + args[0] + "\r\n")
}

func get(args []string) []byte {
	if len(args) != 1 {
		return []byte("-ERR wrong number of arguments for 'get' command\r\n")
	}
	key := args[0]
	store.RLock()
	defer store.RUnlock()
	if value, exists := store.m[key]; exists {
		return []byte("$" + fmt.Sprintf("%d", len(value)) + "\r\n" + value + "\r\n")
	} else {
		return []byte("$-1\r\n") // Null bulk string for non-existent key
	}
}

func set(args []string) []byte {
	if len(args) != 2 {
		return []byte("-ERR wrong number of arguments for 'set' command\r\n")
	}
	key, value := args[0], args[1]
	store.Lock()
	defer store.Unlock()
	store.m[key] = value
	return []byte("+OK\r\n")
}

func respBodyWrapper(args []string) (b []byte) {
	b = append(b, []byte("*"+strconv.Itoa(len(args))+"\r\n")...)
	for _, arg := range args {
		if len(arg) == 0 {
			b = append(b, []byte("$-1\r\n")...) // Null bulk string
			continue
		}
		b = append(b, []byte("$"+strconv.Itoa(len(arg))+"\r\n"+arg+"\r\n")...)
	}
	return b
}
