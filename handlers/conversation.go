package handlers

import "sync"

// convStep represents the current step in a multi-step conversation.
type convStep int

const (
	convStepNone      convStep = iota
	convStepNBPDate            // waiting for the income/expense date input
	convStepNBPAmount          // waiting for the amount input (date already stored)
)

// convState holds the in-progress state for a single chat's conversation.
// States are stored per chat ID and cleaned up after the flow completes.
//
// Note: This is an in-memory store. State is lost on bot restart, which is
// acceptable for a stateless Cloud Run deployment — users simply start over.
type convState struct {
	Step convStep
	Date string // set after convStepNBPDate is completed
}

var (
	convMu    sync.RWMutex
	convStore = make(map[int64]*convState)
)

// getConvState returns the active conversation state for the given chat, or nil.
func getConvState(chatID int64) *convState {
	convMu.RLock()
	defer convMu.RUnlock()
	return convStore[chatID]
}

// setConvState saves or updates the conversation state for the given chat.
func setConvState(chatID int64, s *convState) {
	convMu.Lock()
	defer convMu.Unlock()
	convStore[chatID] = s
}

// clearConvState removes the conversation state for the given chat.
func clearConvState(chatID int64) {
	convMu.Lock()
	defer convMu.Unlock()
	delete(convStore, chatID)
}
