package utils

import "github.com/google/uuid"

func GenerateUID() string {
	uuidV4 := uuid.New()
	return uuidV4.String()
}

func GenerateHeartbeatID() string {
	uuidV4 := uuid.New()
	return "heartbeat-" + uuidV4.String()
}
