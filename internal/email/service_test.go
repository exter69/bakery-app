package email

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogSender_DoesNotError(t *testing.T) {
	// Arrange
	sender := &LogSender{}
	msg := EmailMessage{
		To:      "customer@example.com",
		Subject: "Order Confirmed",
		Body:    "<h1>Thank you!</h1>",
	}

	// Act
	err := sender.Send(context.Background(), msg)

	// Assert
	require.NoError(t, err)
}

func TestSMTPSender_StoresConfig(t *testing.T) {
	// Arrange & Act
	sender := &SMTPSender{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "user@example.com",
		Password: "secret",
		From:     "noreply@example.com",
	}

	// Assert
	assert.Equal(t, "smtp.example.com", sender.Host)
	assert.Equal(t, 587, sender.Port)
	assert.Equal(t, "user@example.com", sender.Username)
	assert.Equal(t, "secret", sender.Password)
	assert.Equal(t, "noreply@example.com", sender.From)
}

func TestLogSender_ImplementsSenderInterface(t *testing.T) {
	// Assert that LogSender satisfies the Sender interface at compile time
	var _ Sender = (*LogSender)(nil)
	var _ Sender = (*SMTPSender)(nil)
}
