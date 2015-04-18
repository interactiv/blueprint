package main

// Distributed application that exposes the polls as a rest api

import (
	"net/http"
	"sync"
)

var (
	vars     map[*http.Request]map[string]interface{}
	varslock sync.RWMutex
)

// OpenVars prepares the vars map
func OpenVars(r *http.Request) {
	varslock.Lock()
	if vars == nil {
		vars = make(map[*http.Request]map[string]interface{}{})
	}
	vars[r] = map[string]interface{}{}
	varslock.Unlock()
}

// Safely delete an entry in the vars map
func ClosVars(r *http.Request) {
	varslock.Lock()
	delete(vars, r)
	varslock.Unlock()
}

// Get a value from vars
func GetVar(r *http.Request, key string) interface{} {
	varslock.RLock()
	value := vars[r][key]
	varslock.RUnlock()
	return value
}

// Set a value to vars
func SetVar(r *http.Request, key string, value interface{}) {
	varslock.Lock()
	vars[r][key] = value
	varslock.Unlock()
}
