package producer

import (
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/orochi-keydream/counter-service/internal/config"
	"github.com/orochi-keydream/counter-service/internal/model"
)

type DialogueCommandProducer struct {
	producer sarama.SyncProducer
	topic    string
}

func NewDialogueCommandProducer(config config.KafkaConfig) (*DialogueCommandProducer, error) {
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true

	syncProducer, err := sarama.NewSyncProducer(config.Brokers, cfg)

	if err != nil {
		return nil, err
	}

	producer := &DialogueCommandProducer{
		producer: syncProducer,
		topic:    config.Producers.DialogueCommands.Topic,
	}

	return producer, nil
}

func (p *DialogueCommandProducer) SendMessage(message *model.OutboxMessage) error {
	var (
		messageKeyBytes   []byte
		messageValueBytes []byte
	)

	switch message.Type {
	case model.OutboxMessageTypeRollbackMessage:
		messageKey := message.MessageKey.(string)
		messageValue := message.MessageValue.(model.RollbackMessage)

		messageKeyBytes = []byte(messageKey)

		bytes, err := mapRollbackMessageToBytes(messageValue)

		if err != nil {
			return err
		}

		messageValueBytes = bytes
	case model.OutboxMessageTypeCommitMessage:
		messageKey := message.MessageKey.(string)
		messageValue := message.MessageValue.(model.CommitMessage)

		messageKeyBytes = []byte(messageKey)

		bytes, err := mapCommitMessageToBytes(messageValue)

		if err != nil {
			return err
		}

		messageValueBytes = bytes
	default:
		return fmt.Errorf("Unsupported message type: %s", message.Type)
	}

	msg := &sarama.ProducerMessage{
		Key:   sarama.StringEncoder(messageKeyBytes),
		Value: sarama.StringEncoder(messageValueBytes),
		Topic: p.topic,
	}

	_, _, err := p.producer.SendMessage(msg)

	return err
}

type messageCommand string

type message struct {
	CorrelationId string          `json:"correlationId"`
	Command       messageCommand  `json:"command"`
	Payload       json.RawMessage `json:"payload"`
}

const (
	MessageCommandCommitMessage   messageCommand = "CommitMessage"
	MessageCommandRollbackMessage messageCommand = "RollbackMessage"
)

type commitMessagePayload struct {
	MessageId int64 `json:"messageId"`
}

type rollbackMessagePayload struct {
	MessageId int64 `json:"messageId"`
}

func mapRollbackMessageToBytes(msg model.RollbackMessage) ([]byte, error) {
	payload := rollbackMessagePayload{
		MessageId: int64(msg.MessageId),
	}

	payloadBytes, err := json.Marshal(payload)

	if err != nil {
		return nil, err
	}

	m := message{
		CorrelationId: string(msg.CorrelationId),
		Command:       MessageCommandRollbackMessage,
		Payload:       payloadBytes,
	}

	return json.Marshal(m)
}

func mapCommitMessageToBytes(msg model.CommitMessage) ([]byte, error) {
	payload := commitMessagePayload{
		MessageId: int64(msg.MessageId),
	}

	payloadBytes, err := json.Marshal(payload)

	if err != nil {
		return nil, err
	}

	m := message{
		CorrelationId: string(msg.CorrelationId),
		Command:       MessageCommandCommitMessage,
		Payload:       payloadBytes,
	}

	return json.Marshal(m)
}
