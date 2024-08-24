package model

// Value objects

type MessageId int64

type UserId string

type ChatId string

type CorrelationId string

// Models

type Message struct {
	MessageId MessageId
	UserId    UserId
	ChatId    ChatId
	IsRead    bool
}

type OutboxMessageType int32

const (
	OutboxMessageTypeRollbackMessage = 1
	OutboxMessageTypeCommitMessage   = 2
)

type OutboxMessage struct {
	Id           int64
	Type         OutboxMessageType
	MessageKey   any
	MessageValue any
	IsSent       bool
}

type RollbackMessage struct {
	CorrelationId CorrelationId
	MessageId     MessageId
}

type CommitMessage struct {
	CorrelationId CorrelationId
	MessageId     MessageId
}
