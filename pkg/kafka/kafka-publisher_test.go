package kafka

import (
	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestSetSASLParameters_AllEnvVarsSet(t *testing.T) {
	err := os.Setenv("PUBLISH_TOPIC_USERNAME", "testuser")
	assert.NoError(t, err)
	err = os.Setenv("PUBLISH_TOPIC_PASSWORD", "testpassword")
	assert.NoError(t, err)
	err = os.Setenv("PUBLISH_TOPIC_MECHANISM", "SCRAM-SHA-512")
	assert.NoError(t, err)
	defer func() {
		err = os.Unsetenv("PUBLISH_TOPIC_USERNAME")
		assert.NoError(t, err)
		err = os.Unsetenv("PUBLISH_TOPIC_PASSWORD")
		assert.NoError(t, err)
		err = os.Unsetenv("PUBLISH_TOPIC_MECHANISM")
		assert.NoError(t, err)
	}()

	config := sarama.NewConfig()
	setSASLParameters(config) // rename if package-level symbol differs

	assert.True(t, config.Net.SASL.Enable)
	assert.Equal(t, "testuser", config.Net.SASL.User)
	assert.Equal(t, "testpassword", config.Net.SASL.Password)
	assert.Equal(t, sarama.SASLMechanism("SCRAM-SHA-512"), config.Net.SASL.Mechanism)
	assert.True(t, config.Net.SASL.Handshake)
}

func TestSetSASLParameters_MechanismDefault(t *testing.T) {
	err := os.Setenv("PUBLISH_TOPIC_USERNAME", "testuser")
	assert.NoError(t, err)
	err = os.Setenv("PUBLISH_TOPIC_PASSWORD", "testpassword")
	assert.NoError(t, err)
	err = os.Unsetenv("PUBLISH_TOPIC_MECHANISM")
	assert.NoError(t, err)

	config := sarama.NewConfig()
	setSASLParameters(config)

	assert.Equal(t, "testuser", config.Net.SASL.User)
	assert.Equal(t, "testpassword", config.Net.SASL.Password)
	assert.Equal(t, sarama.SASLMechanism("SCRAM-SHA-512"), config.Net.SASL.Mechanism)
}
