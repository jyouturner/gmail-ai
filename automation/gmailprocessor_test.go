package automation

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/jyouturer/gmail-ai/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/api/gmail/v1"
)

func TestMain(m *testing.M) {
	logger, err := logging.NewLogger()
	if err != nil {
		panic(err)
	}
	logging.Logger = logger // Set the global logger instance
}

type mockEmailService struct {
	mock.Mock
}

func (m *mockEmailService) GetHistoryList(userId string, startHistoryId uint64) (uint64, *gmail.ListHistoryResponse, error) {
	args := m.Called(userId, startHistoryId)
	return args.Get(0).(uint64), args.Get(1).(*gmail.ListHistoryResponse), args.Error(2)
}

func (m *mockEmailService) GetMessage(userId string, id string) (*gmail.Message, error) {
	args := m.Called(userId, id)
	return args.Get(0).(*gmail.Message), args.Error(1)
}

func TestEmailProcessor(t *testing.T) {
	mockService := &mockEmailService{}
	handlers := []EmailHandlerFunc{
		func(ctx context.Context, email *gmail.Message) error {
			return errors.New("Error processing email")
		},
	}

	mockService.On("GetMessage", "me", "1234").Return(&gmail.Message{}, nil)
	mockService.On("GetMessage", "me", "5678").Return(nil, errors.New("Error retrieving message"))
	mockService.On("GetHistoryList", "me", uint64(0)).Return(uint64(5678), &gmail.ListHistoryResponse{
		History: []*gmail.History{
			{
				Id: 1234,
				MessagesAdded: []*gmail.HistoryMessageAdded{
					{
						Message: &gmail.Message{
							Id: "1234",
							Payload: &gmail.MessagePart{
								Headers: []*gmail.MessagePartHeader{
									{
										Name:  "To",
										Value: "test@example.com",
									},
									{
										Name:  "From",
										Value: "noreply@example.com",
									},
								},
							},
						},
					},
				},
			},
		},
	}, nil)

	mockPollHistory := &mockPollHistory{}
	mockPollHistory.On("ReadHistory").Return(uint64(0), nil)
	mockPollHistory.On("WriteHistory", uint64(1234)).Return(nil)

	ep := &EmailProvider{
		service: mockService,
	}
	err := ep.PollAndProcess(context.Background(), mockPollHistory, handlers)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Error processing email")

	err = ep.PollAndProcess(context.Background(), mockPollHistory, handlers)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Error retrieving message")
}

type mockPollHistory struct {
	mock.Mock
}

func (m *mockPollHistory) ReadHistory() (uint64, error) {
	args := m.Called()
	return args.Get(0).(uint64), args.Error(1)
}

func (m *mockPollHistory) WriteHistory(historyId uint64) error {
	args := m.Called(historyId)
	return args.Error(0)
}

func TestFileHistory_ReadHistoryId(t *testing.T) {
	f, err := ioutil.TempFile("", "test")
	assert.Nil(t, err)
	defer os.Remove(f.Name())

	filename := f.Name()
	f.Write([]byte("1234"))

	history := &FileHistory{
		filename: filename,
	}
	historyId, err := history.ReadHistory()
	assert.Nil(t, err)
	assert.Equal(t, uint64(1234), historyId)
}

func TestFileHistory_WriteHistoryId(t *testing.T) {
	dir, err := ioutil.TempDir("", "history-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	filename := filepath.Join(dir, "history.txt")
	fh := &FileHistory{filename: filename}

	// Write history to file
	historyId := uint64(1234)
	if err := fh.WriteHistory(historyId); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify file contents
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if string(data) != "1234" {
		t.Errorf("unexpected file content: %s", string(data))
	}
}

func TestEmailProvider_PollAndProcess(t *testing.T) {
	// Create a mock email service
	mockService := &mockEmailService{}

	// Create a mock poll history
	mockPollHistory := &mockPollHistory{}

	// Define email handler functions
	var handlers []EmailHandlerFunc
	handler1 := func(ctx context.Context, email *gmail.Message) error { return nil }
	handler2 := func(ctx context.Context, email *gmail.Message) error { return fmt.Errorf("handler2 error") }
	handlers = append(handlers, handler1, handler2)

	// Create an email provider
	ep := &EmailProvider{service: mockService}

	// Define expected values
	expectedHandler1Calls := 2
	expectedHandler2Calls := 1
	expectedHistoryId := uint64(456)

	// Set up mock expectations
	mockService.On("GetHistoryList", "me", uint64(0)).Return(uint64(5678), &gmail.ListHistoryResponse{
		History: []*gmail.History{
			{
				Id: 1234,
				MessagesAdded: []*gmail.HistoryMessageAdded{
					{
						Message: &gmail.Message{
							Id: "1234",
							Payload: &gmail.MessagePart{
								Headers: []*gmail.MessagePartHeader{
									{
										Name:  "To",
										Value: "test@example.com",
									},
									{
										Name:  "From",
										Value: "noreply@example.com",
									},
								},
							},
						},
					},
				},
			},
		},
	}, nil)
	mockService.On("GetMessage", "me", "1234").Return(&gmail.Message{}, nil)
	mockService.On("GetMessage", "me", "5678").Return(nil, errors.New("Error retrieving message"))
	mockService.On("Handler1", mock.Anything, mock.Anything).Return(nil).Times(expectedHandler1Calls)
	mockService.On("Handler2", mock.Anything, mock.Anything).Return(fmt.Errorf("handler2 error")).Times(expectedHandler2Calls)

	mockPollHistory.On("ReadHistory").Return(uint64(0), nil)
	mockPollHistory.On("WriteHistory", expectedHistoryId).Return(nil)

	// Call PollAndProcess with mock service and poll history
	err := ep.PollAndProcess(context.Background(), mockPollHistory, handlers)
	assert.NotNil(t, err)

	// Verify mock expectations were met
	mockService.AssertExpectations(t)
	mockPollHistory.AssertExpectations(t)
}
