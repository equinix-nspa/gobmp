package kafka

type Config struct {
	ServerAddress        string
	TopicRetentionTimeMs string
	SASLEnable           bool
	SASLMechanism        string
	SASLUsername         string
	SASLPassword         string
}
