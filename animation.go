package vtable

import (
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// AnimationManager manages active animations and refresh triggers
type AnimationManager struct {
	mu           sync.RWMutex
	animations   map[string]*Animation
	config       AnimationConfig
	timers       map[string]*time.Timer
	callbacks    map[string]func()
	batchUpdates []string
	batchTimer   *time.Timer
	onRefresh    func([]string) tea.Cmd
}

// NewAnimationManager creates a new animation manager
func NewAnimationManager(config AnimationConfig) *AnimationManager {
	return &AnimationManager{
		animations: make(map[string]*Animation),
		config:     config,
		timers:     make(map[string]*time.Timer),
		callbacks:  make(map[string]func()),
		onRefresh:  func([]string) tea.Cmd { return nil },
	}
}

// SetRefreshCallback sets the callback function for when refreshes are needed
func (am *AnimationManager) SetRefreshCallback(callback func([]string) tea.Cmd) {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.onRefresh = callback
}

// RegisterAnimation registers a new animation with its triggers
func (am *AnimationManager) RegisterAnimation(id string, triggers []RefreshTrigger, state map[string]any) {
	if !am.config.Enabled {
		return
	}

	am.mu.Lock()
	defer am.mu.Unlock()

	// Limit the number of active animations
	if len(am.animations) >= am.config.MaxAnimations {
		// Remove oldest animation
		am.removeOldestAnimation()
	}

	// Create new animation
	animation := &Animation{
		State:      state,
		Triggers:   triggers,
		LastRender: time.Now(),
		IsVisible:  true,
	}

	am.animations[id] = animation
	am.setupTriggers(id, triggers)
}

// UnregisterAnimation removes an animation and cleans up its triggers
func (am *AnimationManager) UnregisterAnimation(id string) {
	am.mu.Lock()
	defer am.mu.Unlock()

	// Clean up timers
	if timer, exists := am.timers[id]; exists {
		timer.Stop()
		delete(am.timers, id)
	}

	// Remove callback
	delete(am.callbacks, id)

	// Remove animation
	delete(am.animations, id)
}

// GetAnimationState returns the current state for an animation
func (am *AnimationManager) GetAnimationState(id string) map[string]any {
	am.mu.RLock()
	defer am.mu.RUnlock()

	if animation, exists := am.animations[id]; exists {
		return animation.State
	}
	return make(map[string]any)
}

// UpdateAnimationState updates the state for an animation
func (am *AnimationManager) UpdateAnimationState(id string, state map[string]any) {
	am.mu.Lock()
	defer am.mu.Unlock()

	if animation, exists := am.animations[id]; exists {
		animation.State = state
		animation.LastRender = time.Now()
	}
}

// SetVisible marks an animation as visible or hidden
func (am *AnimationManager) SetVisible(id string, visible bool) {
	am.mu.Lock()
	defer am.mu.Unlock()

	if animation, exists := am.animations[id]; exists {
		animation.IsVisible = visible
	}
}

// TriggerRefresh manually triggers a refresh for specific animations
func (am *AnimationManager) TriggerRefresh(ids []string) tea.Cmd {
	if am.config.BatchUpdates {
		am.mu.Lock()
		am.batchUpdates = append(am.batchUpdates, ids...)

		// Set up batch timer if not already set
		if am.batchTimer == nil {
			am.batchTimer = time.AfterFunc(16*time.Millisecond, func() {
				am.processBatchUpdates()
			})
		}
		am.mu.Unlock()
		return nil
	}

	return am.onRefresh(ids)
}

// ProcessEvent processes an event that might trigger animations
func (am *AnimationManager) ProcessEvent(event string) tea.Cmd {
	if !am.config.Enabled {
		return nil
	}

	am.mu.RLock()
	var triggeredIDs []string

	for id, animation := range am.animations {
		if !animation.IsVisible {
			continue
		}

		for _, trigger := range animation.Triggers {
			if trigger.Type == TriggerEvent && trigger.Event == event {
				triggeredIDs = append(triggeredIDs, id)
				break
			}
		}
	}
	am.mu.RUnlock()

	if len(triggeredIDs) > 0 {
		return am.TriggerRefresh(triggeredIDs)
	}

	return nil
}

// CheckConditionalTriggers checks all conditional triggers
func (am *AnimationManager) CheckConditionalTriggers() tea.Cmd {
	if !am.config.Enabled {
		return nil
	}

	am.mu.RLock()
	var triggeredIDs []string

	for id, animation := range am.animations {
		if !animation.IsVisible {
			continue
		}

		for _, trigger := range animation.Triggers {
			if trigger.Type == TriggerConditional && trigger.Condition != nil && trigger.Condition() {
				triggeredIDs = append(triggeredIDs, id)
				break
			}
		}
	}
	am.mu.RUnlock()

	if len(triggeredIDs) > 0 {
		return am.TriggerRefresh(triggeredIDs)
	}

	return nil
}

// Cleanup removes all animations and stops all timers
func (am *AnimationManager) Cleanup() {
	am.mu.Lock()
	defer am.mu.Unlock()

	// Stop all timers
	for _, timer := range am.timers {
		timer.Stop()
	}
	am.timers = make(map[string]*time.Timer)

	// Clear callbacks
	am.callbacks = make(map[string]func())

	// Clear animations
	am.animations = make(map[string]*Animation)

	// Stop batch timer
	if am.batchTimer != nil {
		am.batchTimer.Stop()
		am.batchTimer = nil
	}
}

// setupTriggers sets up timers and callbacks for animation triggers
func (am *AnimationManager) setupTriggers(id string, triggers []RefreshTrigger) {
	for _, trigger := range triggers {
		switch trigger.Type {
		case TriggerTimer:
			if trigger.Interval > 0 {
				am.setupTimerTrigger(id, trigger.Interval)
			}
		case TriggerOnce:
			// Trigger once after initial render
			go func(animID string) {
				time.Sleep(time.Millisecond) // Allow initial render to complete
				am.TriggerRefresh([]string{animID})
			}(id)
		}
	}
}

// setupTimerTrigger creates a repeating timer for an animation
func (am *AnimationManager) setupTimerTrigger(id string, interval time.Duration) {
	// Stop existing timer if any
	if timer, exists := am.timers[id]; exists {
		timer.Stop()
	}

	// Respect reduced motion preference
	actualInterval := interval
	if am.config.ReducedMotion {
		actualInterval = interval * 2 // Slow down animations
	}

	// Create new timer
	timer := time.AfterFunc(actualInterval, func() {
		am.mu.RLock()
		animation, exists := am.animations[id]
		am.mu.RUnlock()

		if exists && animation.IsVisible {
			am.TriggerRefresh([]string{id})
			// Reschedule
			am.setupTimerTrigger(id, interval)
		}
	})

	am.timers[id] = timer
}

// removeOldestAnimation removes the animation that was last rendered longest ago
func (am *AnimationManager) removeOldestAnimation() {
	var oldestID string
	var oldestTime time.Time

	for id, animation := range am.animations {
		if oldestID == "" || animation.LastRender.Before(oldestTime) {
			oldestID = id
			oldestTime = animation.LastRender
		}
	}

	if oldestID != "" {
		// Clean up the oldest animation
		if timer, exists := am.timers[oldestID]; exists {
			timer.Stop()
			delete(am.timers, oldestID)
		}
		delete(am.callbacks, oldestID)
		delete(am.animations, oldestID)
	}
}

// processBatchUpdates processes accumulated batch updates
func (am *AnimationManager) processBatchUpdates() {
	am.mu.Lock()
	updates := make([]string, len(am.batchUpdates))
	copy(updates, am.batchUpdates)
	am.batchUpdates = am.batchUpdates[:0]
	am.batchTimer = nil
	am.mu.Unlock()

	if len(updates) > 0 {
		// Remove duplicates
		uniqueUpdates := make(map[string]bool)
		var finalUpdates []string
		for _, update := range updates {
			if !uniqueUpdates[update] {
				uniqueUpdates[update] = true
				finalUpdates = append(finalUpdates, update)
			}
		}

		if cmd := am.onRefresh(finalUpdates); cmd != nil {
			// In a real implementation, this would be sent through the tea.Program
			// For now, we'll just ignore the command
			_ = cmd
		}
	}
}

// GetActiveAnimations returns the IDs of all active animations
func (am *AnimationManager) GetActiveAnimations() []string {
	am.mu.RLock()
	defer am.mu.RUnlock()

	var ids []string
	for id, animation := range am.animations {
		if animation.IsVisible {
			ids = append(ids, id)
		}
	}
	return ids
}

// GetConfig returns the current animation configuration
func (am *AnimationManager) GetConfig() AnimationConfig {
	am.mu.RLock()
	defer am.mu.RUnlock()
	return am.config
}

// UpdateConfig updates the animation configuration
func (am *AnimationManager) UpdateConfig(config AnimationConfig) {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.config = config

	// If animations are disabled, clean up
	if !config.Enabled {
		// We can't call Cleanup() here as it would deadlock
		// Instead, mark all animations as invisible
		for _, animation := range am.animations {
			animation.IsVisible = false
		}
	}
}
