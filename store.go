package main

import "sync"

var store = struct {
	sync.RWMutex
	m map[string]string
}{
	RWMutex: sync.RWMutex{},
	m:       make(map[string]string),
}
