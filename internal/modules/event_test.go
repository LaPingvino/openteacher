package modules

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/LaPingvino/openteacher/internal/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEvent(t *testing.T) {
	t.Run("creation", func(t *testing.T) {
		event := NewEvent("test-event")
		assert.Equal(t, "test-event", event.Name())
	})

	t.Run("subscribe_and_trigger", func(t *testing.T) {
		event := NewEvent("test-event")
		var receivedData interface{}
		var callCount int

		handler := func(data interface{}) error {
			receivedData = data
			callCount++
			return nil
		}

		err := event.Subscribe(handler)
		require.NoError(t, err)

		testData := "test data"
		err = event.Trigger(testData)
		require.NoError(t, err)

		assert.Equal(t, testData, receivedData)
		assert.Equal(t, 1, callCount)
	})

	t.Run("subscribe_nil_handler", func(t *testing.T) {
		event := NewEvent("test-event")
		err := event.Subscribe(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "handler cannot be nil")
	})

	t.Run("unsubscribe_nil_handler", func(t *testing.T) {
		event := NewEvent("test-event")
		err := event.Unsubscribe(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "handler cannot be nil")
	})

	t.Run("multiple_handlers", func(t *testing.T) {
		event := NewEvent("test-event")
		var handler1Called, handler2Called bool

		handler1 := func(data interface{}) error {
			handler1Called = true
			return nil
		}

		handler2 := func(data interface{}) error {
			handler2Called = true
			return nil
		}

		err := event.Subscribe(handler1)
		require.NoError(t, err)
		err = event.Subscribe(handler2)
		require.NoError(t, err)

		err = event.Trigger("test")
		require.NoError(t, err)

		assert.True(t, handler1Called)
		assert.True(t, handler2Called)
	})

	t.Run("handler_error", func(t *testing.T) {
		event := NewEvent("test-event")
		expectedError := errors.New("handler error")

		handler := func(data interface{}) error {
			return expectedError
		}

		err := event.Subscribe(handler)
		require.NoError(t, err)

		err = event.Trigger("test")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "handler error")
	})

	t.Run("unsubscribe", func(t *testing.T) {
		event := NewEvent("test-event")
		var callCount int

		handler := func(data interface{}) error {
			callCount++
			return nil
		}

		err := event.Subscribe(handler)
		require.NoError(t, err)

		err = event.Trigger("test1")
		require.NoError(t, err)
		assert.Equal(t, 1, callCount)

		err = event.Unsubscribe(handler)
		require.NoError(t, err)

		err = event.Trigger("test2")
		require.NoError(t, err)
		assert.Equal(t, 1, callCount) // Should not have increased
	})

	t.Run("unsubscribe_nonexistent_handler", func(t *testing.T) {
		event := NewEvent("test-event")

		handler := func(data interface{}) error {
			return nil
		}

		err := event.Unsubscribe(handler)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "handler not found")
	})
}

func TestEventModule(t *testing.T) {
	t.Run("creation", func(t *testing.T) {
		module := NewEventModule()

		assert.Equal(t, "event", module.Type())
		assert.Equal(t, "event-module", module.Name())
		assert.Equal(t, 2000, module.Priority())
		assert.False(t, module.IsActive())
		assert.Equal(t, 0, module.EventCount())
	})

	t.Run("lifecycle", func(t *testing.T) {
		module := NewEventModule()
		ctx := context.Background()

		// Enable module
		err := module.Enable(ctx)
		require.NoError(t, err)
		assert.True(t, module.IsActive())

		// Disable module
		err = module.Disable(ctx)
		require.NoError(t, err)
		assert.False(t, module.IsActive())
		assert.Equal(t, 0, module.EventCount()) // Events should be cleared
	})

	t.Run("create_event", func(t *testing.T) {
		module := NewEventModule()

		event := module.CreateEvent("test-event")
		assert.NotNil(t, event)
		assert.Equal(t, "test-event", event.Name())
		assert.Equal(t, 1, module.EventCount())

		// Creating same event should return existing one
		event2 := module.CreateEvent("test-event")
		assert.Equal(t, event, event2)
		assert.Equal(t, 1, module.EventCount()) // Should not increase
	})

	t.Run("create_event_empty_name", func(t *testing.T) {
		module := NewEventModule()

		assert.Panics(t, func() {
			module.CreateEvent("")
		})
	})

	t.Run("get_event", func(t *testing.T) {
		module := NewEventModule()

		// Non-existent event
		event, exists := module.GetEvent("non-existent")
		assert.Nil(t, event)
		assert.False(t, exists)

		// Create and get event
		createdEvent := module.CreateEvent("test-event")
		event, exists = module.GetEvent("test-event")
		assert.Equal(t, createdEvent, event)
		assert.True(t, exists)
	})

	t.Run("list_events", func(t *testing.T) {
		module := NewEventModule()

		// Initially empty
		events := module.ListEvents()
		assert.Empty(t, events)

		// Add some events
		module.CreateEvent("event1")
		module.CreateEvent("event2")
		module.CreateEvent("event3")

		events = module.ListEvents()
		assert.Len(t, events, 3)
		assert.Contains(t, events, "event1")
		assert.Contains(t, events, "event2")
		assert.Contains(t, events, "event3")
	})

	t.Run("subscribe_and_trigger", func(t *testing.T) {
		module := NewEventModule()
		var receivedData interface{}
		var callCount int

		handler := func(data interface{}) error {
			receivedData = data
			callCount++
			return nil
		}

		// Subscribe to non-existent event (should create it)
		err := module.Subscribe("test-event", handler)
		require.NoError(t, err)
		assert.Equal(t, 1, module.EventCount())

		// Trigger the event
		testData := "test data"
		err = module.Trigger("test-event", testData)
		require.NoError(t, err)

		assert.Equal(t, testData, receivedData)
		assert.Equal(t, 1, callCount)
	})

	t.Run("subscribe_empty_event_name", func(t *testing.T) {
		module := NewEventModule()

		handler := func(data interface{}) error { return nil }

		err := module.Subscribe("", handler)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "event name cannot be empty")
	})

	t.Run("subscribe_nil_handler", func(t *testing.T) {
		module := NewEventModule()

		err := module.Subscribe("test-event", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "handler cannot be nil")
	})

	t.Run("unsubscribe", func(t *testing.T) {
		module := NewEventModule()
		var callCount int

		handler := func(data interface{}) error {
			callCount++
			return nil
		}

		// Subscribe first
		err := module.Subscribe("test-event", handler)
		require.NoError(t, err)

		// Trigger once
		err = module.Trigger("test-event", "test1")
		require.NoError(t, err)
		assert.Equal(t, 1, callCount)

		// Unsubscribe
		err = module.Unsubscribe("test-event", handler)
		require.NoError(t, err)

		// Trigger again - should not call handler
		err = module.Trigger("test-event", "test2")
		require.NoError(t, err)
		assert.Equal(t, 1, callCount) // Should not have increased
	})

	t.Run("unsubscribe_empty_event_name", func(t *testing.T) {
		module := NewEventModule()

		handler := func(data interface{}) error { return nil }

		err := module.Unsubscribe("", handler)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "event name cannot be empty")
	})

	t.Run("unsubscribe_nil_handler", func(t *testing.T) {
		module := NewEventModule()

		err := module.Unsubscribe("test-event", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "handler cannot be nil")
	})

	t.Run("unsubscribe_nonexistent_event", func(t *testing.T) {
		module := NewEventModule()

		handler := func(data interface{}) error { return nil }

		err := module.Unsubscribe("nonexistent-event", handler)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "event \"nonexistent-event\" not found")
	})

	t.Run("trigger_empty_event_name", func(t *testing.T) {
		module := NewEventModule()

		err := module.Trigger("", "data")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "event name cannot be empty")
	})

	t.Run("trigger_nonexistent_event", func(t *testing.T) {
		module := NewEventModule()

		// Should not error - just no handlers to call
		err := module.Trigger("nonexistent-event", "data")
		assert.NoError(t, err)
	})

	t.Run("interface_compliance", func(t *testing.T) {
		module := NewEventModule()

		// Should implement Module interface
		var _ core.Module = module

		// Should implement EventModule interface
		var _ core.EventModule = module
	})

	t.Run("disable_clears_events", func(t *testing.T) {
		module := NewEventModule()
		ctx := context.Background()

		err := module.Enable(ctx)
		require.NoError(t, err)

		// Create some events
		module.CreateEvent("event1")
		module.CreateEvent("event2")
		assert.Equal(t, 2, module.EventCount())

		// Disable should clear events
		err = module.Disable(ctx)
		require.NoError(t, err)
		assert.Equal(t, 0, module.EventCount())
	})
}

func TestEventModuleConcurrency(t *testing.T) {
	t.Run("concurrent_event_creation", func(t *testing.T) {
		module := NewEventModule()
		const numGoroutines = 10
		const eventsPerGoroutine = 5

		var wg sync.WaitGroup
		wg.Add(numGoroutines)

		// Create events concurrently
		for i := 0; i < numGoroutines; i++ {
			go func(routineID int) {
				defer wg.Done()
				for j := 0; j < eventsPerGoroutine; j++ {
					eventName := fmt.Sprintf("event-%d-%d", routineID, j)
					event := module.CreateEvent(eventName)
					assert.NotNil(t, event)
				}
			}(i)
		}

		wg.Wait()

		// Should have created all events
		expectedCount := numGoroutines * eventsPerGoroutine
		assert.Equal(t, expectedCount, module.EventCount())
	})

	t.Run("concurrent_subscribe_and_trigger", func(t *testing.T) {
		module := NewEventModule()
		const numSubscribers = 5
		const numTriggers = 3

		// Counters for each subscriber
		counters := make([]int, numSubscribers)
		var counterMutex sync.Mutex

		var wg sync.WaitGroup

		// Start subscribers
		for i := 0; i < numSubscribers; i++ {
			wg.Add(1)
			go func(subscriberID int) {
				defer wg.Done()
				handler := func(data interface{}) error {
					counterMutex.Lock()
					counters[subscriberID]++
					counterMutex.Unlock()
					return nil
				}

				err := module.Subscribe("test-event", handler)
				assert.NoError(t, err)
			}(i)
		}

		// Start triggers
		for i := 0; i < numTriggers; i++ {
			wg.Add(1)
			go func(triggerID int) {
				defer wg.Done()
				// Small delay to let some subscribers register
				time.Sleep(10 * time.Millisecond)

				err := module.Trigger("test-event", fmt.Sprintf("trigger-%d", triggerID))
				assert.NoError(t, err)
			}(i)
		}

		wg.Wait()

		// Each subscriber should have received some events
		// (Exact count may vary due to timing, but should be > 0 for most)
		counterMutex.Lock()
		totalEvents := 0
		for _, count := range counters {
			totalEvents += count
		}
		counterMutex.Unlock()

		// Should have received some events overall
		assert.Greater(t, totalEvents, 0)
	})

	t.Run("concurrent_list_and_create", func(t *testing.T) {
		module := NewEventModule()
		const numCreators = 3
		const numListers = 2
		const eventsPerCreator = 5

		var wg sync.WaitGroup

		// Start event creators
		for i := 0; i < numCreators; i++ {
			wg.Add(1)
			go func(creatorID int) {
				defer wg.Done()
				for j := 0; j < eventsPerCreator; j++ {
					eventName := fmt.Sprintf("event-%d-%d", creatorID, j)
					module.CreateEvent(eventName)
				}
			}(i)
		}

		// Start event listers (should not panic or deadlock)
		for i := 0; i < numListers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for k := 0; k < 10; k++ {
					events := module.ListEvents()
					assert.NotNil(t, events)
					time.Sleep(1 * time.Millisecond)
				}
			}()
		}

		wg.Wait()

		// Final count should be correct
		expectedCount := numCreators * eventsPerCreator
		assert.Equal(t, expectedCount, module.EventCount())
	})
}
