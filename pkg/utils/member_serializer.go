package utils

import (
	"encoding/json"
	"fmt"

	tb "gopkg.in/telebot.v4"
)

func SerializeMember(member *tb.ChatMember) ([]byte, error) {
	data, err := json.Marshal(member)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize member: %w", err)
	}
	return data, nil
}

func DeserializeMember(data []byte) (*tb.ChatMember, error) {
	var member tb.ChatMember
	if err := json.Unmarshal(data, &member); err != nil {
		return nil, fmt.Errorf("failed to deserialize member: %w", err)
	}
	return &member, nil
}
