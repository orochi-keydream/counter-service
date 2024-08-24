package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/IBM/sarama"
	"github.com/orochi-keydream/counter-service/internal/model"
	"github.com/orochi-keydream/counter-service/internal/service"
	"log"
	"sync"
)

type CounterCommandsConsumer struct {
	counterService *service.CounterService
}

func NewCounterCommandsConsumer(counterService *service.CounterService) *CounterCommandsConsumer {
	return &CounterCommandsConsumer{counterService: counterService}
}

func RunCounterCommandsConsumer(ctx context.Context, addrs []string, topic string, c *CounterCommandsConsumer, wg *sync.WaitGroup) error {
	cfg := sarama.NewConfig()
	cfg.Consumer.Offsets.Initial = sarama.OffsetOldest

	cg, err := sarama.NewConsumerGroup(addrs, "counter-service", cfg)

	if err != nil {
		panic(err)
	}

	topics := []string{topic}

	go func() {
		defer wg.Done()

		for {
			err = cg.Consume(ctx, topics, c)

			if err != nil {
				if errors.Is(err, sarama.ErrClosedConsumerGroup) {
					return
				}

				log.Panicln("Something wrong with the consumer")
			}

			if ctx.Err() != nil {
				return
			}
		}
	}()

	return nil
}

func (c *CounterCommandsConsumer) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (c *CounterCommandsConsumer) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (c *CounterCommandsConsumer) ConsumeClaim(cgs sarama.ConsumerGroupSession, cgc sarama.ConsumerGroupClaim) error {
	for {
		select {
		case msg, ok := <-cgc.Messages():
			if !ok {
				log.Println("Message channel was closed")
				return nil
			}

			log.Printf("Handling message with offset %v from %v topic\n", msg.Offset, msg.Topic)

			err := c.handle(msg.Value, cgs.Context())

			if err != nil {
				log.Println(err)
				continue
			}

			cgs.MarkMessage(msg, "")
		case <-cgs.Context().Done():
			log.Println("ConsumeClaim: cancellation requested")
			return nil
		}
	}
}

func (c *CounterCommandsConsumer) handle(msgBytes []byte, ctx context.Context) error {
	var msg counterCommandMessage
	err := json.Unmarshal(msgBytes, &msg)

	if err != nil {
		return err
	}

	cmd := model.AddNewUnreadMessageCommand{
		CorrelationId: model.CorrelationId(msg.CorrelationId),
		UserId:        model.UserId(msg.UserId),
		ChatId:        model.ChatId(msg.ChatId),
		MessageId:     model.MessageId(msg.MessageId),
	}

	return c.counterService.AddNewUnreadMessage(ctx, cmd)
}

type counterCommandMessage struct {
	CorrelationId string `json:"correlationId"`
	UserId        string `json:"userId"`
	ChatId        string `json:"chatId"`
	MessageId     int64  `json:"messageId"`
}
