package configs

// Configuration ... The configuration of server, database
type Configuration struct {
    Server          string
    ServerPort      string
    AWSRegion       string
    DynamoDBServer  string
    DynamoDBPort    string
    DynamoDBTable   string
}