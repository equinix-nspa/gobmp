package kafka

import (
	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetSASLParameters_AllFieldsSet(t *testing.T) {
	kConfig := &Config{
		SASLUsername:  "testuser",
		SASLPassword:  "testpassword",
		SASLMechanism: "PLAIN",
	}
	config := sarama.NewConfig()

	setSASLParameters(config, kConfig)

	assert.True(t, config.Net.SASL.Enable)
	assert.Equal(t, "testuser", config.Net.SASL.User)
	assert.Equal(t, "testpassword", config.Net.SASL.Password)
	assert.Equal(t, sarama.SASLMechanism("PLAIN"), config.Net.SASL.Mechanism)
	assert.True(t, config.Net.SASL.Handshake)
}

func TestSetSASLParameters_DefaultMechanism(t *testing.T) {
	kConfig := &Config{
		SASLUsername: "testuser",
		SASLPassword: "testpassword",
	}
	config := sarama.NewConfig()

	setSASLParameters(config, kConfig)

	assert.True(t, config.Net.SASL.Enable)
	assert.Equal(t, "testuser", config.Net.SASL.User)
	assert.Equal(t, "testpassword", config.Net.SASL.Password)
	assert.Equal(t, sarama.SASLTypeSCRAMSHA512, string(config.Net.SASL.Mechanism))
	assert.NotNil(t, config.Net.SASL.SCRAMClientGeneratorFunc)
	assert.True(t, config.Net.SASL.Handshake)
}

func TestSetSASLParameters_EmptyConfig(t *testing.T) {
	kConfig := &Config{}
	config := sarama.NewConfig()

	setSASLParameters(config, kConfig)

	assert.True(t, config.Net.SASL.Enable)
	assert.Empty(t, config.Net.SASL.User)
	assert.Empty(t, config.Net.SASL.Password)
	assert.Equal(t, sarama.SASLTypeSCRAMSHA512, string(config.Net.SASL.Mechanism))
	assert.NotNil(t, config.Net.SASL.SCRAMClientGeneratorFunc)
	assert.True(t, config.Net.SASL.Handshake)
}
