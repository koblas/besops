package types

import (
	"context"
	"fmt"
	"strings"
	"time"

	pahomqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/koblas/besops/internal/monitor"
	"github.com/koblas/besops/lib/status"
)

type MQTTChecker struct{}

func (c *MQTTChecker) Type() string { return "mqtt" }

func (c *MQTTChecker) Check(ctx context.Context, cfg *monitor.Config) (monitor.CheckResult, error) {
	if cfg.MQTT.Topic == "" {
		return monitor.CheckResult{
			Status:  status.Down,
			Message: "MQTT topic is required",
		}, nil
	}

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = time.Duration(cfg.Interval.Seconds()*0.8) * time.Second
		if timeout == 0 {
			timeout = 16 * time.Second
		}
	}

	broker := cfg.Hostname
	if !strings.Contains(broker, "://") {
		broker = "tcp://" + broker
	}
	if cfg.Port > 0 {
		broker = fmt.Sprintf("%s:%d", broker, cfg.Port)
	}

	messageCh := make(chan mqttMessage, 1)
	errCh := make(chan error, 1)

	opts := pahomqtt.NewClientOptions().
		AddBroker(broker).
		SetConnectTimeout(timeout).
		SetWriteTimeout(timeout).
		SetOnConnectHandler(func(client pahomqtt.Client) {
			token := client.Subscribe(cfg.MQTT.Topic, 0, func(_ pahomqtt.Client, msg pahomqtt.Message) {
				messageCh <- mqttMessage{
					topic:   msg.Topic(),
					payload: string(msg.Payload()),
				}
			})
			if !token.WaitTimeout(timeout) {
				errCh <- fmt.Errorf("subscribe timeout")
			} else if token.Error() != nil {
				errCh <- fmt.Errorf("subscribe failed: %w", token.Error())
			}
		}).
		SetConnectionLostHandler(func(_ pahomqtt.Client, err error) {
			errCh <- fmt.Errorf("connection lost: %w", err)
		})

	if cfg.MQTT.Username != "" {
		opts.SetUsername(cfg.MQTT.Username)
	}
	if cfg.MQTT.Password != "" {
		opts.SetPassword(cfg.MQTT.Password)
	}

	start := time.Now()
	client := pahomqtt.NewClient(opts)
	token := client.Connect()
	if !token.WaitTimeout(timeout) {
		return monitor.CheckResult{
			Status:  status.Down,
			Message: "MQTT connection timeout",
		}, nil
	}
	if token.Error() != nil {
		return monitor.CheckResult{
			Status:  status.Down,
			Message: fmt.Sprintf("MQTT connection failed: %v", token.Error()),
		}, nil
	}
	defer client.Disconnect(250)

	select {
	case <-ctx.Done():
		return monitor.CheckResult{
			Status:  status.Down,
			Message: "check cancelled",
		}, nil
	case err := <-errCh:
		ping := time.Since(start).Milliseconds()
		return monitor.CheckResult{
			Status:  status.Down,
			Ping:    ping,
			Message: err.Error(),
		}, nil
	case msg := <-messageCh:
		ping := time.Since(start).Milliseconds()
		return c.evaluateMessage(cfg, msg, ping)
	case <-time.After(timeout):
		ping := time.Since(start).Milliseconds()
		return monitor.CheckResult{
			Status:  status.Down,
			Ping:    ping,
			Message: "timeout waiting for MQTT message",
		}, nil
	}
}

func (c *MQTTChecker) evaluateMessage(cfg *monitor.Config, msg mqttMessage, ping int64) (monitor.CheckResult, error) {
	if cfg.MQTT.SuccessMessage == "" {
		return monitor.CheckResult{
			Status:  status.Up,
			Ping:    ping,
			Message: fmt.Sprintf("Topic: %s; Message: %s", msg.topic, msg.payload),
		}, nil
	}

	if strings.Contains(msg.payload, cfg.MQTT.SuccessMessage) {
		return monitor.CheckResult{
			Status:  status.Up,
			Ping:    ping,
			Message: fmt.Sprintf("Topic: %s; Message: %s", msg.topic, msg.payload),
		}, nil
	}

	return monitor.CheckResult{
		Status:  status.Down,
		Ping:    ping,
		Message: fmt.Sprintf("message mismatch - Topic: %s; Message: %s", msg.topic, msg.payload),
	}, nil
}

type mqttMessage struct {
	topic   string
	payload string
}
