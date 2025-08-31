package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/uaru-shit/votes/internal/domain"
)

// implements PollStorage interface using JSON files
type FilePollStorage struct {
	filePath string
	mutex    sync.RWMutex
}

func NewFilePollStorage(filePath string) (*FilePollStorage, error) {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	storage := &FilePollStorage{
		filePath: filePath,
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if err := storage.savePollsToFile([]*domain.ActivePoll{}); err != nil {
			return nil, fmt.Errorf("failed to initialize storage file: %w", err)
		}
	}

	return storage, nil
}

func (s *FilePollStorage) SavePoll(poll *domain.ActivePoll) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	polls, err := s.loadPollsFromFile()
	if err != nil {
		return fmt.Errorf("failed to load polls: %w", err)
	}

	found := false
	for i, existingPoll := range polls {
		if existingPoll.ID == poll.ID {
			polls[i] = poll
			found = true
			break
		}
	}

	if !found {
		polls = append(polls, poll)
	}

	return s.savePollsToFile(polls)
}

func (s *FilePollStorage) GetPolls() ([]*domain.ActivePoll, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.loadPollsFromFile()
}

func (s *FilePollStorage) DeletePoll(id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	polls, err := s.loadPollsFromFile()
	if err != nil {
		return fmt.Errorf("failed to load polls: %w", err)
	}

	var filteredPolls []*domain.ActivePoll
	for _, poll := range polls {
		if poll.ID != id {
			filteredPolls = append(filteredPolls, poll)
		}
	}

	return s.savePollsToFile(filteredPolls)
}

func (s *FilePollStorage) GetPollsByType(pollType domain.PollType) ([]*domain.ActivePoll, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	polls, err := s.loadPollsFromFile()
	if err != nil {
		return nil, fmt.Errorf("failed to load polls: %w", err)
	}

	var filteredPolls []*domain.ActivePoll
	for _, poll := range polls {
		if poll.Type == pollType {
			filteredPolls = append(filteredPolls, poll)
		}
	}

	return filteredPolls, nil
}

func (s *FilePollStorage) CleanExpiredPolls() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	polls, err := s.loadPollsFromFile()
	if err != nil {
		return fmt.Errorf("failed to load polls: %w", err)
	}

	var activePolls []*domain.ActivePoll
	now := time.Now()

	for _, poll := range polls {
		if poll.ExpiresAt.After(now) {
			activePolls = append(activePolls, poll)
		}
	}

	return s.savePollsToFile(activePolls)
}

func (s *FilePollStorage) loadPollsFromFile() ([]*domain.ActivePoll, error) {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var polls []*domain.ActivePoll
	if len(data) > 0 {
		if err := json.Unmarshal(data, &polls); err != nil {
			return nil, fmt.Errorf("failed to unmarshal polls: %w", err)
		}
	}

	return polls, nil
}

func (s *FilePollStorage) savePollsToFile(polls []*domain.ActivePoll) error {
	data, err := json.MarshalIndent(polls, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal polls: %w", err)
	}

	if err := os.WriteFile(s.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
