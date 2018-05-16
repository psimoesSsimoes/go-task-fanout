package logger

import "testing"

func TestSpawnMute(t *testing.T) {
	SpawnMute()
	SpawnSimpleMute()
	SpawnCompatibleMute()
}
