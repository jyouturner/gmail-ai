package messagesource

import (
	"fmt"

	"github.com/jyouturer/gmail-ai/datamodel"
	"github.com/jyouturer/gmail-ai/integration"
	"github.com/jyouturer/gmail-ai/internal/logging"
	"go.uber.org/zap"
	"google.golang.org/api/gmail/v1"
)

type GmailService struct {
	Gmail *gmail.Service
}

func NewGmailService(gmail *gmail.Service) *GmailService {
	return &GmailService{Gmail: gmail}
}

// GetMessages retrieves message by ID.
func (s *GmailService) GetMessage(userId, messageID string) (datamodel.Message, error) {
	m, err := s.Gmail.Users.Messages.Get(userId, messageID).Format("full").Do()
	if err != nil {
		return datamodel.Message{}, fmt.Errorf("failed to get message: %v", err)
	}
	return s.FromGmailMessage(m), nil
}

// FromGmailMessage converts a gmail.Message to datamodel.Message
func (s *GmailService) FromGmailMessage(m *gmail.Message) datamodel.Message {
	return datamodel.Message{
		ID:      m.Id,
		Subject: m.Snippet,
		From:    m.Payload.Headers[0].Value,
		To:      m.Payload.Headers[1].Value,
		Body:    s.GetBody(m),
	}
}

// GetMessageIds retrieves message ids by history id.
func (s *GmailService) GetMessageIds(userId string, startHistoryId uint64) (uint64, []string, error) {
	// first get the history list
	lastHistoryId, histories, err := s.GetHistoryList(userId, startHistoryId)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to retrieve history: %v", err)
	}
	// iterate and get message ids
	var ids []string
	for _, h := range histories.History {
		lastHistoryId = h.Id
		for _, m := range h.MessagesAdded {
			messageID := m.Message.Id
			ids = append(ids, messageID)
		}
	}
	return lastHistoryId, ids, nil
}

// GetMessages retrieves messages by IDs
func (s *GmailService) GetMessages(userId string, ids []string) ([]datamodel.Message, error) {
	var messages []datamodel.Message
	for _, id := range ids {
		msg, err := s.GetMessage("me", id)
		if err != nil {
			continue
		}
		messages = append(messages, msg)
	}
	return messages, nil
}

// GetHistoryList retrieves history list by the starting history id.
func (s *GmailService) GetHistoryList(userId string, startHistoryId uint64) (uint64, *gmail.ListHistoryResponse, error) {
	var nextPageToken string
	var histories []*gmail.History

	for {
		req := s.Gmail.Users.History.List(userId).StartHistoryId(startHistoryId).MaxResults(500)
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

// GetBody retrieves the body of the message by extracting the critical comoonents using NLP
func (s *GmailService) GetBody(msg *gmail.Message) string {
	// Get the text of the email
	text, err := integration.GetMessageCriticalContents(msg)
	if err != nil {
		logging.Logger.Warn("error in parsing message, will use snippet", zap.String("id", msg.Id), zap.Error(err))
		text = msg.Snippet
	}
	if text == "" {
		logging.Logger.Warn("empty message from parsing, will use snippet", zap.String("id", msg.Id))
		text = msg.Snippet
	}
	return text
}
