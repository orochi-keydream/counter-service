package model

type GetUnreadCountTotalCommand struct {
	UserId UserId
}

type GetUnreadCountCommand struct {
	CurrentUserId UserId
	ChatUserId    UserId
}

type MarkMessagesAsReadCommand struct {
	UserId     UserId
	MessageIds []MessageId
}

type AddNewUnreadMessageCommand struct {
	CorrelationId CorrelationId
	UserId        UserId
	ChatId        ChatId
	MessageId     MessageId
}
