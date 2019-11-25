package configs

import "os"

import "encoding/json"

// Configuration ... The configuration of server, database
type Configuration struct {
	Server         string
	ServerPort     string
	AWSRegion      string
	DynamoDBServer string
	DynamoDBPort   string
	DynamoDBTable  string
	BaseDN         string
	Filter         string
	ROUserName     string
	ROUserPass     string
	Host           string
	LDAPEnable     string
}

// ObtainConfig ... The function obtain config from a json file
func ObtainConfig(filename string) (config Configuration, err error) {
	file, _ := os.Open(filename)
	defer file.Close()
	decoder := json.NewDecoder(file)
	config = Configuration{}
	err = decoder.Decode(&config)
	return config, err
}
