package configs

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
