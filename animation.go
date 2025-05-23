package vtable

import (
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Global animation ticker message - sent continuously
type GlobalAnimationTickMsg struct {
	Timestamp time.Time
}

// Animation update message - sent when animations actually change
type AnimationUpdateMsg struct {
	UpdatedAnimations []string
}

// AnimationState holds the current state of an animation
type AnimationState struct {
	ID         string
	State      map[string]any
	Triggers   []RefreshTrigger
	LastUpdate time.Time
	NextUpdate time.Time
	IsActive   bool
	IsVisible  bool
	IsDirty    bool // Whether the animation has changed since last render
}

// AnimationEngine manages animations using a single global timer
type AnimationEngine struct {
	mu               sync.RWMutex
	animations       map[string]*AnimationState
	config           AnimationConfig
	needsUpdate      bool
	lastGlobalUpdate time.Time
}

// NewAnimationEngine creates a new animation engine
func NewAnimationEngine(config AnimationConfig) *AnimationEngine {
	engine := &AnimationEngine{
		animations:       make(map[string]*AnimationState),
		config:           config,
		lastGlobalUpdate: time.Now(),
	}

	return engine
}

// startGlobalTicker creates the global animation ticker
func (ae *AnimationEngine) startGlobalTicker() tea.Cmd {
	// Use configurable tick interval from config
	tickInterval := ae.config.TickInterval
	if tickInterval <= 0 {
		tickInterval = 100 * time.Millisecond // Fallback to reasonable default
	}

	if ae.config.ReducedMotion {
		tickInterval = tickInterval * 2 // Double the interval for reduced motion
	}

	return tea.Tick(tickInterval, func(t time.Time) tea.Msg {
		return GlobalAnimationTickMsg{Timestamp: t}
	})
}

// ProcessGlobalTick processes the global animation tick
func (ae *AnimationEngine) ProcessGlobalTick(msg GlobalAnimationTickMsg) tea.Cmd {
	if !ae.config.Enabled {
		return nil
	}

	ae.mu.Lock()
	defer ae.mu.Unlock()

	now := msg.Timestamp
	updatedAnimations := []string{}
	hasUpdates := false

	// Check each animation to see if it needs updating
	for id, animation := range ae.animations {
		if !animation.IsActive || !animation.IsVisible {
			continue
		}

		// Check if any timer triggers are due
		shouldUpdate := false
		for _, trigger := range animation.Triggers {
			if trigger.Type == TriggerTimer && now.After(animation.NextUpdate) {
				shouldUpdate = true
				// Schedule next update
				animation.NextUpdate = now.Add(trigger.Interval)
				break
			}
		}

		if shouldUpdate {
			animation.LastUpdate = now
			animation.IsDirty = true
			updatedAnimations = append(updatedAnimations, id)
			hasUpdates = true
		}
	}

	// Create batch commands
	var cmds []tea.Cmd

	// Always schedule the next global tick to maintain the animation loop
	cmds = append(cmds, ae.startGlobalTicker())

	// If we have updates, send an update message
	if hasUpdates {
		ae.needsUpdate = true
		ae.lastGlobalUpdate = now
		cmds = append(cmds, func() tea.Msg {
			return AnimationUpdateMsg{UpdatedAnimations: updatedAnimations}
		})
	}

	return tea.Batch(cmds...)
}

// RegisterAnimation registers a new animation
func (ae *AnimationEngine) RegisterAnimation(id string, triggers []RefreshTrigger, initialState map[string]any) tea.Cmd {
	if !ae.config.Enabled {
		return nil
	}

	ae.mu.Lock()
	defer ae.mu.Unlock()

	// Limit active animations
	if len(ae.animations) >= ae.config.MaxAnimations {
		ae.removeOldestAnimationUnsafe()
	}

	now := time.Now()

	// Create animation state
	animation := &AnimationState{
		ID:         id,
		State:      make(map[string]any),
		Triggers:   triggers,
		LastUpdate: now,
		NextUpdate: now, // Will be updated based on triggers
		IsActive:   true,
		IsVisible:  true,
		IsDirty:    true, // New animations are dirty by default
	}

	// Copy initial state
	for k, v := range initialState {
		animation.State[k] = v
	}

	// Set next update time based on timer triggers
	for _, trigger := range triggers {
		if trigger.Type == TriggerTimer && trigger.Interval > 0 {
			animation.NextUpdate = now.Add(trigger.Interval)
			break
		}
	}

	ae.animations[id] = animation

	// Start global ticker if this is the first animation
	if len(ae.animations) == 1 {
		return ae.startGlobalTicker()
	}

	return nil
}

// UnregisterAnimation removes an animation
func (ae *AnimationEngine) UnregisterAnimation(id string) tea.Cmd {
	ae.mu.Lock()
	defer ae.mu.Unlock()

	if animation, exists := ae.animations[id]; exists {
		animation.IsActive = false
		delete(ae.animations, id)
	}

	return nil
}

// GetAnimationState returns the current state for an animation
func (ae *AnimationEngine) GetAnimationState(id string) map[string]any {
	ae.mu.RLock()
	defer ae.mu.RUnlock()

	if animation, exists := ae.animations[id]; exists && animation.IsActive {
		// Return a copy to prevent race conditions
		stateCopy := make(map[string]any)
		for k, v := range animation.State {
			stateCopy[k] = v
		}
		return stateCopy
	}

	return make(map[string]any)
}

// UpdateAnimationState updates the state for an animation
func (ae *AnimationEngine) UpdateAnimationState(id string, newState map[string]any) {
	ae.mu.Lock()
	defer ae.mu.Unlock()

	if animation, exists := ae.animations[id]; exists && animation.IsActive {
		// Update state
		hasChanges := false
		for k, v := range newState {
			if animation.State[k] != v {
				animation.State[k] = v
				hasChanges = true
			}
		}

		if hasChanges {
			animation.LastUpdate = time.Now()
			animation.IsDirty = true
		}
	}
}

// SetVisible marks an animation as visible or hidden
func (ae *AnimationEngine) SetVisible(id string, visible bool) {
	ae.mu.Lock()
	defer ae.mu.Unlock()

	if animation, exists := ae.animations[id]; exists {
		if animation.IsVisible != visible {
			animation.IsVisible = visible
			animation.IsDirty = true
		}
	}
}

// IsVisible returns whether an animation is visible
func (ae *AnimationEngine) IsVisible(id string) bool {
	ae.mu.RLock()
	defer ae.mu.RUnlock()

	if animation, exists := ae.animations[id]; exists {
		return animation.IsVisible && animation.IsActive
	}
	return false
}

// HasUpdates checks if any animations have been updated since last check
func (ae *AnimationEngine) HasUpdates() bool {
	ae.mu.RLock()
	defer ae.mu.RUnlock()

	for _, animation := range ae.animations {
		if animation.IsActive && animation.IsVisible && animation.IsDirty {
			return true
		}
	}

	return false
}

// ClearDirtyFlags clears the dirty flags for all animations
func (ae *AnimationEngine) ClearDirtyFlags() {
	ae.mu.Lock()
	defer ae.mu.Unlock()

	for _, animation := range ae.animations {
		animation.IsDirty = false
	}
}

// GetActiveAnimations returns the IDs of all active visible animations
func (ae *AnimationEngine) GetActiveAnimations() []string {
	ae.mu.RLock()
	defer ae.mu.RUnlock()

	var ids []string
	for id, animation := range ae.animations {
		if animation.IsActive && animation.IsVisible {
			ids = append(ids, id)
		}
	}
	return ids
}

// GetDirtyAnimations returns the IDs of animations that need re-rendering
func (ae *AnimationEngine) GetDirtyAnimations() []string {
	ae.mu.RLock()
	defer ae.mu.RUnlock()

	var ids []string
	for id, animation := range ae.animations {
		if animation.IsActive && animation.IsVisible && animation.IsDirty {
			ids = append(ids, id)
		}
	}
	return ids
}

// Cleanup removes all animations
func (ae *AnimationEngine) Cleanup() {
	ae.mu.Lock()
	defer ae.mu.Unlock()

	// Mark all animations as inactive
	for _, animation := range ae.animations {
		animation.IsActive = false
	}

	// Clear all maps
	ae.animations = make(map[string]*AnimationState)
	ae.needsUpdate = false
}

// GetConfig returns the current configuration
func (ae *AnimationEngine) GetConfig() AnimationConfig {
	ae.mu.RLock()
	defer ae.mu.RUnlock()
	return ae.config
}

// UpdateConfig updates the configuration
func (ae *AnimationEngine) UpdateConfig(config AnimationConfig) {
	ae.mu.Lock()
	defer ae.mu.Unlock()
	ae.config = config

	if !config.Enabled {
		// Disable all animations
		for _, animation := range ae.animations {
			animation.IsActive = false
		}
	}
}

// ProcessEvent processes external events that might trigger animations
func (ae *AnimationEngine) ProcessEvent(event string) []string {
	ae.mu.RLock()
	defer ae.mu.RUnlock()

	var triggeredIDs []string
	for id, animation := range ae.animations {
		if !animation.IsActive || !animation.IsVisible {
			continue
		}

		for _, trigger := range animation.Triggers {
			if trigger.Type == TriggerEvent && trigger.Event == event {
				triggeredIDs = append(triggeredIDs, id)
				break
			}
		}
	}

	return triggeredIDs
}

// CheckConditionalTriggers checks all conditional triggers and returns triggered IDs
func (ae *AnimationEngine) CheckConditionalTriggers() []string {
	ae.mu.RLock()
	defer ae.mu.RUnlock()

	var triggeredIDs []string
	for id, animation := range ae.animations {
		if !animation.IsActive || !animation.IsVisible {
			continue
		}

		for _, trigger := range animation.Triggers {
			if trigger.Type == TriggerConditional && trigger.Condition != nil && trigger.Condition() {
				triggeredIDs = append(triggeredIDs, id)
				break
			}
		}
	}

	return triggeredIDs
}

// removeOldestAnimationUnsafe removes the oldest animation (must be called with lock held)
func (ae *AnimationEngine) removeOldestAnimationUnsafe() {
	var oldestID string
	var oldestTime time.Time

	for id, animation := range ae.animations {
		if oldestID == "" || animation.LastUpdate.Before(oldestTime) {
			oldestID = id
			oldestTime = animation.LastUpdate
		}
	}

	if oldestID != "" {
		delete(ae.animations, oldestID)
	}
}

// Global animation engine instance
var globalAnimationEngine *AnimationEngine

// InitializeAnimationEngine initializes the global animation engine
func InitializeAnimationEngine(config AnimationConfig) {
	globalAnimationEngine = NewAnimationEngine(config)
}

// GetAnimationEngine returns the global animation engine
func GetAnimationEngine() *AnimationEngine {
	if globalAnimationEngine == nil {
		globalAnimationEngine = NewAnimationEngine(DefaultAnimationConfig())
	}
	return globalAnimationEngine
}

// StartGlobalAnimationLoop starts the global animation loop
func StartGlobalAnimationLoop() tea.Cmd {
	engine := GetAnimationEngine()
	if engine.config.Enabled {
		return engine.startGlobalTicker()
	}
	return nil
}

// Deprecated: AnimationManager - use AnimationEngine instead
type AnimationManager = AnimationEngine

// Deprecated: NewAnimationManager - use NewAnimationEngine instead
func NewAnimationManager(config AnimationConfig) *AnimationManager {
	return NewAnimationEngine(config)
}

// Deprecated: AnimationTickMsg, AnimationStartMsg, AnimationStopMsg - use GlobalAnimationTickMsg and AnimationUpdateMsg instead
type AnimationTickMsg struct {
	AnimationID string
	Timestamp   time.Time
}

type AnimationStartMsg struct {
	AnimationID string
}

type AnimationStopMsg struct {
	AnimationID string
}
