package service

type RealtimeConfigService struct {
	isAlwaysRollbackMessagesEnabled bool
}

func NewRealtimeConfigService() *RealtimeConfigService {
	return &RealtimeConfigService{
		isAlwaysRollbackMessagesEnabled: false,
	}
}

func (s *RealtimeConfigService) IsAlwaysRollbackMessagesEnabled() bool {
	return s.isAlwaysRollbackMessagesEnabled
}

func (s *RealtimeConfigService) SetIsAlwaysRollbackMessagesEnabled(value bool) {
	s.isAlwaysRollbackMessagesEnabled = value
}
