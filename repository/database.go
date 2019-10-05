package repository

//Config - Interface to map of string config values
type Config interface {
	Category() string
	Values() map[string]string
}

//BasicConfig - Basic implementation of Config Map with a backing map and convenience add method
type BasicConfig struct {
	cat  string
	vals map[string]string
}

//Category - Returns the category name of this config (debug purposes only foreseen right now, hence no ID)
func (config *BasicConfig) Category() string {
	return config.cat
}

//Values - Returns the valuemap
func (config *BasicConfig) Values() map[string]string {
	return config.vals
}

//AddEntry - Adds a new map entry
func (config *BasicConfig) AddEntry(key string, value string) {
	config.vals[key] = value
}

//NewBasicConfig - Returns a basic implementation of a Config
func NewBasicConfig(category string) *BasicConfig {
	config := BasicConfig{cat: category, vals: make(map[string]string)}
	return &config
}

//Repository - Interface to access table CRUD methods, insert, update, etc...; Contains Config and Implements Session for DB types to reuse
type Repository interface {
	Config() Config
	SetSession(Session)
	Session
}

//Session - Interface to enable strategy to be set in library for communicating to DB...
type Session interface {
	Session() Session
}
