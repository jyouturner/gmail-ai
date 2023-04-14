package automation

import (
	"fmt"

	"google.golang.org/api/gmail/v1"
)

type GmailService struct {
	service *gmail.Service
}

func NewGmailService(service *gmail.Service) *GmailService {
	return &GmailService{service: service}
}

func (s *GmailService) GetMessage(userId, messageID string) (*gmail.Message, error) {
	return s.service.Users.Messages.Get(userId, messageID).Format("full").Do()
}

func (s *GmailService) GetHistoryList(userId string, startHistoryId uint64) (uint64, *gmail.ListHistoryResponse, error) {
	var nextPageToken string
	var histories []*gmail.History

	for {
		req := s.service.Users.History.List(userId).StartHistoryId(startHistoryId).MaxResults(500)
		if nextPageToken != "" {
			req = req.PageToken(nextPageToken)
		}

		resp, err := req.Do()
		if err != nil {
			return 0, nil, fmt.Errorf("failed to retrieve history: %v", err)
		}

		histories = append(histories, resp.History...)
		if len(resp.History) == 0 {
			break
		}

		startHistoryId = resp.History[len(resp.History)-1].Id
		nextPageToken = resp.NextPageToken
	}

	return startHistoryId, &gmail.ListHistoryResponse{History: histories}, nil
}
