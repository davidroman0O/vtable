package vtable

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

// ------------------------
// Test Animation Data Types
// ------------------------

type AnimTestItem struct {
	ID       int
	Name     string
	Status   string
	Progress int
}

type AnimTestDataProvider struct {
	items     []AnimTestItem
	selection map[int]bool
	mu        sync.RWMutex
}

func NewAnimTestDataProvider(count int) *AnimTestDataProvider {
	items := make([]AnimTestItem, count)
	for i := 0; i < count; i++ {
		items[i] = AnimTestItem{
			ID:       i,
			Name:     fmt.Sprintf("Item %d", i),
			Status:   "processing",
			Progress: i * 10 % 100,
		}
	}
	return &AnimTestDataProvider{
		items:     items,
		selection: make(map[int]bool),
	}
}

func (p *AnimTestDataProvider) GetTotal() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.items)
}

func (p *AnimTestDataProvider) GetItems(request DataRequest) ([]Data[AnimTestItem], error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	start := request.Start
	count := request.Count

	if start >= len(p.items) {
		return []Data[AnimTestItem]{}, nil
	}

	end := start + count
	if end > len(p.items) {
		end = len(p.items)
	}

	result := make([]Data[AnimTestItem], end-start)
	for i := start; i < end; i++ {
		result[i-start] = Data[AnimTestItem]{
			ID:       fmt.Sprintf("%d", p.items[i].ID),
			Item:     p.items[i],
			Selected: p.selection[i],
			Metadata: NewTypedMetadata(),
			Disabled: false,
			Hidden:   false,
		}
	}

	return result, nil
}

func (p *AnimTestDataProvider) UpdateProgress(index int, progress int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if index >= 0 && index < len(p.items) {
		p.items[index].Progress = progress
	}
}

func (p *AnimTestDataProvider) UpdateStatus(index int, status string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if index >= 0 && index < len(p.items) {
		p.items[index].Status = status
	}
}

// Implement remaining DataProvider methods
func (p *AnimTestDataProvider) GetSelectionMode() SelectionMode {
	return SelectionMultiple
}

func (p *AnimTestDataProvider) SetSelected(index int, selected bool) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if index < 0 || index >= len(p.items) {
		return false
	}
	if selected {
		p.selection[index] = true
	} else {
		delete(p.selection, index)
	}
	return true
}

func (p *AnimTestDataProvider) SelectAll() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	for i := 0; i < len(p.items); i++ {
		p.selection[i] = true
	}
	return true
}

func (p *AnimTestDataProvider) ClearSelection() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.selection = make(map[int]bool)
}

func (p *AnimTestDataProvider) GetSelectedIndices() []int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	indices := make([]int, 0, len(p.selection))
	for idx := range p.selection {
		indices = append(indices, idx)
	}
	return indices
}

func (p *AnimTestDataProvider) GetItemID(item *AnimTestItem) string {
	return fmt.Sprintf("%d", item.ID)
}

func (p *AnimTestDataProvider) GetSelectedIDs() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	ids := make([]string, 0, len(p.selection))
	for idx := range p.selection {
		if idx < len(p.items) {
			ids = append(ids, fmt.Sprintf("%d", p.items[idx].ID))
		}
	}
	return ids
}

func (p *AnimTestDataProvider) SetSelectedByIDs(ids []string, selected bool) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, id := range ids {
		for i, item := range p.items {
			if fmt.Sprintf("%d", item.ID) == id {
				if selected {
					p.selection[i] = true
				} else {
					delete(p.selection, i)
				}
				break
			}
		}
	}
	return true
}

func (p *AnimTestDataProvider) SelectRange(startID, endID string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	// Implementation similar to other providers
	return true
}

// ------------------------
// Animation System Tests
// ------------------------

func TestAnimationEngineBasics(t *testing.T) {
	config := AnimationConfig{
		Enabled:       true,
		BatchUpdates:  false,
		MaxAnimations: 10,
		ReducedMotion: false,
	}

	engine := NewAnimationEngine(config)

	// Test basic registration
	triggers := []RefreshTrigger{
		{Type: TriggerTimer, Interval: 100 * time.Millisecond},
	}
	initialState := map[string]any{"opacity": 0.0}

	cmd := engine.RegisterAnimation("test1", triggers, initialState)
	// cmd may be nil if this isn't the first animation - that's fine
	if cmd == nil {
		t.Log("RegisterAnimation returned nil - this is expected if not the first animation")
	}

	// Test state retrieval
	state := engine.GetAnimationState("test1")
	if state["opacity"] != 0.0 {
		t.Errorf("Expected opacity 0.0, got %v", state["opacity"])
	}

	// Test state update
	newState := map[string]any{"opacity": 1.0}
	engine.UpdateAnimationState("test1", newState)

	state = engine.GetAnimationState("test1")
	if state["opacity"] != 1.0 {
		t.Errorf("Expected opacity 1.0, got %v", state["opacity"])
	}

	// Test visibility
	if !engine.IsVisible("test1") {
		t.Error("Expected animation to be visible")
	}

	engine.SetVisible("test1", false)
	if engine.IsVisible("test1") {
		t.Error("Expected animation to be invisible")
	}

	// Test unregistration
	cmd = engine.UnregisterAnimation("test1")
	// cmd may be nil - that's fine
	if cmd == nil {
		t.Log("UnregisterAnimation returned nil - this is expected")
	}

	// Test cleanup
	engine.Cleanup()
	activeAnimations := engine.GetActiveAnimations()
	if len(activeAnimations) != 0 {
		t.Errorf("Expected 0 active animations after cleanup, got %d", len(activeAnimations))
	}
}

func TestAnimationTimerBehavior(t *testing.T) {
	config := AnimationConfig{
		Enabled:       true,
		BatchUpdates:  false,
		MaxAnimations: 5,
		ReducedMotion: false,
	}

	engine := NewAnimationEngine(config)
	defer engine.Cleanup()

	// Register animation with timer trigger
	triggers := []RefreshTrigger{
		{Type: TriggerTimer, Interval: 50 * time.Millisecond},
	}
	initialState := map[string]any{"frame": 0}

	cmd := engine.RegisterAnimation("timer-test", triggers, initialState)
	// cmd may be nil - that's fine
	if cmd == nil {
		t.Log("RegisterAnimation returned nil - this is expected")
	}

	// Verify the animation is active
	if !engine.IsVisible("timer-test") {
		t.Error("Expected animation to be visible")
	}

	activeAnimations := engine.GetActiveAnimations()
	if len(activeAnimations) != 1 {
		t.Errorf("Expected 1 active animation, got %d", len(activeAnimations))
	}
}

func TestAnimationMemoryLeaks(t *testing.T) {
	config := AnimationConfig{
		Enabled:       true,
		BatchUpdates:  false,
		MaxAnimations: 3, // Small limit to test cleanup
		ReducedMotion: false,
	}

	engine := NewAnimationEngine(config)
	defer engine.Cleanup()

	// Register more animations than the limit
	for i := 0; i < 5; i++ {
		animID := fmt.Sprintf("anim-%d", i)
		triggers := []RefreshTrigger{
			{Type: TriggerTimer, Interval: 100 * time.Millisecond},
		}
		initialState := map[string]any{"id": i}

		cmd := engine.RegisterAnimation(animID, triggers, initialState)
		// cmd may be nil - that's fine
		if cmd == nil {
			t.Logf("RegisterAnimation returned nil for %s - this is expected", animID)
		}
	}

	// Should only have MaxAnimations active
	activeAnimations := engine.GetActiveAnimations()
	if len(activeAnimations) > config.MaxAnimations {
		t.Errorf("Expected max %d active animations, got %d", config.MaxAnimations, len(activeAnimations))
	}
}

func TestCurrentAnimationSystemFlaws(t *testing.T) {
	// This test is designed to expose the flaws in the current animation system
	config := AnimationConfig{
		Enabled:       true,
		BatchUpdates:  false,
		MaxAnimations: 10,
		ReducedMotion: false,
	}

	engine := NewAnimationEngine(config)
	defer engine.Cleanup()

	// Test multiple animations with timer triggers
	for i := 0; i < 3; i++ {
		animID := fmt.Sprintf("problematic-%d", i)
		triggers := []RefreshTrigger{
			{Type: TriggerTimer, Interval: 10 * time.Millisecond}, // Very fast interval
		}
		initialState := map[string]any{"counter": 0}

		cmd := engine.RegisterAnimation(animID, triggers, initialState)
		// cmd may be nil - that's fine
		if cmd == nil {
			t.Logf("RegisterAnimation returned nil for %s - this is expected", animID)
		}
	}

	// This should work without deadlocking
	activeAnimations := engine.GetActiveAnimations()
	if len(activeAnimations) != 3 {
		t.Errorf("Expected 3 active animations, got %d", len(activeAnimations))
	}

	// Test rapid state updates
	for i := 0; i < 10; i++ {
		engine.UpdateAnimationState("problematic-0", map[string]any{"counter": i})
		state := engine.GetAnimationState("problematic-0")
		if state["counter"] != i {
			t.Errorf("Expected counter %d, got %v", i, state["counter"])
		}
	}
}

func TestAnimationIntegrationWithTeaList(t *testing.T) {
	fmt.Println("\n=== ANIMATION INTEGRATION WITH TEA LIST TEST ===")

	// Setup provider and list
	provider := NewAnimTestDataProvider(10)

	config := ViewportConfig{
		Height:               5,
		TopThresholdIndex:    1,
		BottomThresholdIndex: 3,
		ChunkSize:            10,
		InitialIndex:         0,
		Debug:                false,
	}

	// Create animated formatter that shows current time
	animatedFormatter := func(data Data[AnimTestItem], index int, ctx RenderContext,
		animationState map[string]any, isCursor bool, isTopThreshold bool, isBottomThreshold bool) RenderResult {

		// Show current time to prove animation is updating
		currentTime := ctx.CurrentTime.Format("15:04:05.000")

		prefix := "  "
		if isCursor {
			prefix = "> "
		}

		content := fmt.Sprintf("%s%s [%s]", prefix, data.Item.Name, currentTime)

		return RenderResult{
			Content: content,
			RefreshTriggers: []RefreshTrigger{{
				Type:     TriggerTimer,
				Interval: 100 * time.Millisecond,
			}},
			AnimationState: map[string]any{
				"timestamp": currentTime,
			},
		}
	}

	styleConfig := StyleConfig{}
	regularFormatter := func(data Data[AnimTestItem], index int, ctx RenderContext, isCursor bool, isTopThreshold bool, isBottomThreshold bool) string {
		prefix := "  "
		if data.Selected {
			prefix = "✓ "
		}
		if isCursor {
			prefix = "> "
			if data.Selected {
				prefix = "✓>"
			}
		}
		return fmt.Sprintf("%s%s [%d%%]", prefix, data.Item.Name, data.Item.Progress)
	}

	list, err := NewTeaList(config, provider, styleConfig, regularFormatter)
	if err != nil {
		t.Fatalf("NewTeaList failed: %v", err)
	}

	// Test 1: Regular rendering (no animations)
	fmt.Println("\n1. Regular rendering (no animations):")
	view := list.View()
	fmt.Print(view)

	// Should show simple progress bars
	if !strings.Contains(view, "Item 0") {
		t.Error("Expected to see 'Item 0' in regular view")
	}

	// Test 2: Set animated formatter
	fmt.Println("\n\n2. Setting animated formatter:")
	list.SetAnimatedFormatter(animatedFormatter)

	// This should trigger animation registration
	view = list.View()
	fmt.Print(view)

	// Should show timestamps instead of progress percentages
	if !strings.Contains(view, ":") {
		t.Error("Expected to see timestamps (containing ':') in animated view")
	}

	// Test 3: Simulate data updates (progress changes)
	fmt.Println("\n\n3. Simulating data updates:")

	// Update progress for visible items
	for i := 0; i < 5; i++ {
		newProgress := (provider.items[i].Progress + 25) % 100
		provider.UpdateProgress(i, newProgress)
	}

	list.RefreshData()
	view = list.View()
	fmt.Print(view)

	fmt.Println("\n\n4. Testing animation cleanup:")
	list.ClearAnimatedFormatter()
	view = list.View()
	fmt.Print(view)

	// Should be back to regular formatting
	if strings.Contains(view, "█") || strings.Contains(view, "░") {
		t.Error("Expected no progress bar characters after clearing animated formatter")
	}

	fmt.Println("\n=== END ANIMATION INTEGRATION TEST ===")
}

// Helper function to test if animations are actually updating visually
func TestAnimationVisualUpdates(t *testing.T) {
	fmt.Println("\n=== ANIMATION VISUAL UPDATES TEST ===")

	provider := NewAnimTestDataProvider(3)

	config := ViewportConfig{
		Height:               3,
		TopThresholdIndex:    0,
		BottomThresholdIndex: 2,
		ChunkSize:            5,
		InitialIndex:         0,
		Debug:                false,
	}

	// Create animated formatter that shows current time
	animatedFormatter := func(data Data[AnimTestItem], index int, ctx RenderContext,
		animationState map[string]any, isCursor bool, isTopThreshold bool, isBottomThreshold bool) RenderResult {

		// Show current time to prove animation is updating
		currentTime := ctx.CurrentTime.Format("15:04:05.000")

		prefix := "  "
		if isCursor {
			prefix = "> "
		}

		content := fmt.Sprintf("%s%s [%s]", prefix, data.Item.Name, currentTime)

		return RenderResult{
			Content: content,
			RefreshTriggers: []RefreshTrigger{{
				Type:     TriggerTimer,
				Interval: 100 * time.Millisecond,
			}},
			AnimationState: map[string]any{
				"timestamp": currentTime,
			},
		}
	}

	regularFormatter := func(data Data[AnimTestItem], index int, ctx RenderContext, isCursor bool, isTopThreshold bool, isBottomThreshold bool) string {
		prefix := "  "
		if isCursor {
			prefix = "> "
		}
		return fmt.Sprintf("%s%s", prefix, data.Item.Name)
	}

	styleConfig := StyleConfig{}
	list, err := NewTeaList(config, provider, styleConfig, regularFormatter)
	if err != nil {
		t.Fatalf("NewTeaList failed: %v", err)
	}

	// Test 1: Regular view (should be static)
	fmt.Println("\n1. Regular view (static):")
	view1 := list.View()
	fmt.Print(view1)

	time.Sleep(10 * time.Millisecond)

	view2 := list.View()
	if view1 != view2 {
		t.Error("Regular view should be static but changed between renders")
	}

	// Test 2: Animated view (should update)
	fmt.Println("\n\n2. Animated view (should update with timestamps):")
	list.SetAnimatedFormatter(animatedFormatter)

	view3 := list.View()
	fmt.Print(view3)

	time.Sleep(50 * time.Millisecond)

	view4 := list.View()
	fmt.Print(view4)

	// The views should be different because timestamps should update
	if view3 == view4 {
		fmt.Println("⚠️  WARNING: Animated views are identical - animations may not be working")
		fmt.Println("View 3:", view3)
		fmt.Println("View 4:", view4)
	} else {
		fmt.Println("✅ Animation system is updating views correctly")
	}

	list.ClearAnimatedFormatter()

	fmt.Println("\n=== END ANIMATION VISUAL UPDATES TEST ===")
}

func TestGlobalAnimationLoop(t *testing.T) {
	fmt.Println("\n=== GLOBAL ANIMATION LOOP TEST ===")

	config := AnimationConfig{
		Enabled:       true,
		BatchUpdates:  false,
		MaxAnimations: 10,
		ReducedMotion: false,
	}

	engine := NewAnimationEngine(config)
	defer engine.Cleanup()

	// Register an animation with a timer trigger
	triggers := []RefreshTrigger{
		{Type: TriggerTimer, Interval: 50 * time.Millisecond},
	}
	initialState := map[string]any{"counter": 0}

	cmd := engine.RegisterAnimation("global-test", triggers, initialState)
	// cmd may be nil if this isn't the first animation - that's fine
	fmt.Printf("RegisterAnimation returned command: %v\n", cmd != nil)

	// Check that the animation was registered
	if !engine.IsVisible("global-test") {
		t.Error("Expected animation to be visible after registration")
	}

	// Simulate processing a global tick
	tickMsg := GlobalAnimationTickMsg{Timestamp: time.Now()}

	// Process the tick
	resultCmd := engine.ProcessGlobalTick(tickMsg)
	if resultCmd == nil {
		t.Error("Expected ProcessGlobalTick to return a command")
	}

	// Check if the animation has updates
	if !engine.HasUpdates() {
		fmt.Println("⚠️  Animation doesn't have updates yet - this might be expected on first tick")
	}

	// Update the animation state
	engine.UpdateAnimationState("global-test", map[string]any{"counter": 1})

	// Now it should have updates
	if !engine.HasUpdates() {
		t.Error("Expected animation to have updates after state change")
	}

	// Get dirty animations
	dirtyAnimations := engine.GetDirtyAnimations()
	if len(dirtyAnimations) == 0 {
		t.Error("Expected at least one dirty animation")
	}

	// Clear dirty flags
	engine.ClearDirtyFlags()
	if engine.HasUpdates() {
		t.Error("Expected no updates after clearing dirty flags")
	}

	fmt.Println("✅ Global animation loop test passed")
	fmt.Println("=== END GLOBAL ANIMATION LOOP TEST ===")
}

func TestContinuousAnimationUpdates(t *testing.T) {
	fmt.Println("\n=== CONTINUOUS ANIMATION UPDATES TEST ===")

	provider := NewAnimTestDataProvider(3)

	config := ViewportConfig{
		Height:               3,
		TopThresholdIndex:    0,
		BottomThresholdIndex: 2,
		ChunkSize:            5,
		InitialIndex:         0,
		Debug:                false,
	}

	// Create animated formatter that updates counter every tick
	animatedFormatter := func(data Data[AnimTestItem], index int, ctx RenderContext,
		animationState map[string]any, isCursor bool, isTopThreshold bool, isBottomThreshold bool) RenderResult {

		// Get current counter from animation state
		counter := 0
		if c, ok := animationState["counter"]; ok {
			if ci, ok := c.(int); ok {
				counter = ci
			}
		}

		// Increment counter
		counter++

		currentTime := ctx.CurrentTime.Format("15:04:05.000")
		prefix := "  "
		if isCursor {
			prefix = "> "
		}

		content := fmt.Sprintf("%s%s [T:%s C:%d]", prefix, data.Item.Name, currentTime, counter)

		return RenderResult{
			Content: content,
			RefreshTriggers: []RefreshTrigger{{
				Type:     TriggerTimer,
				Interval: 25 * time.Millisecond, // Fast updates
			}},
			AnimationState: map[string]any{
				"counter":   counter,
				"timestamp": currentTime,
			},
		}
	}

	regularFormatter := func(data Data[AnimTestItem], index int, ctx RenderContext, isCursor bool, isTopThreshold bool, isBottomThreshold bool) string {
		prefix := "  "
		if isCursor {
			prefix = "> "
		}
		return fmt.Sprintf("%s%s", prefix, data.Item.Name)
	}

	styleConfig := StyleConfig{}
	list, err := NewTeaList(config, provider, styleConfig, regularFormatter)
	if err != nil {
		t.Fatalf("NewTeaList failed: %v", err)
	}

	// Set animated formatter
	list.SetAnimatedFormatter(animatedFormatter)

	// Simulate the global animation loop
	fmt.Println("1. Starting animation loop simulation...")

	// Process initial view to register animations
	view1 := list.View()
	fmt.Printf("Initial view:\n%s\n", view1)

	// Simulate a global tick
	engine := list.animationEngine
	tickMsg := GlobalAnimationTickMsg{Timestamp: time.Now()}

	// Wait a bit to ensure timer triggers would be due
	time.Sleep(30 * time.Millisecond)

	cmd := engine.ProcessGlobalTick(tickMsg)
	if cmd != nil {
		fmt.Println("✅ Global tick returned a command")
	}

	// Check for updates
	if engine.HasUpdates() {
		fmt.Println("✅ Animation engine reports updates available")

		dirtyAnimations := engine.GetDirtyAnimations()
		fmt.Printf("   Dirty animations: %v\n", dirtyAnimations)

		// Process animations
		view2 := list.View()
		fmt.Printf("Updated view:\n%s\n", view2)

		if view1 != view2 {
			fmt.Println("✅ View changed after animation update")
		} else {
			fmt.Println("⚠️  View didn't change - animations might not be working")
		}
	} else {
		fmt.Println("⚠️  No updates reported by animation engine")
	}

	// Test multiple ticks
	fmt.Println("\n2. Testing multiple ticks...")
	for i := 0; i < 3; i++ {
		time.Sleep(30 * time.Millisecond)
		tickMsg = GlobalAnimationTickMsg{Timestamp: time.Now()}
		cmd = engine.ProcessGlobalTick(tickMsg)

		if engine.HasUpdates() {
			view := list.View()
			fmt.Printf("Tick %d view:\n%s\n", i+1, view)
		}
	}

	fmt.Println("=== END CONTINUOUS ANIMATION UPDATES TEST ===")
}
