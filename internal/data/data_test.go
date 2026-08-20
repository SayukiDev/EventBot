package data

import (
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
)

func TestNew(t *testing.T) {
	d := New("/tmp/data.bson")
	if d.path != "/tmp/data.bson" {
		t.Errorf("path = %q, want %q", d.path, "/tmp/data.bson")
	}
	if d.content == nil {
		t.Error("content map is nil, want initialized map")
	}
}

func TestSetCreatesNewContent(t *testing.T) {
	d := New("")
	existed := d.Set("channel-1", func(c *Content) {
		if c.Events == nil {
			t.Error("Events map is nil, want initialized map")
		}
		if c.Category == nil {
			t.Error("Category slice is nil, want initialized slice")
		}
		c.Events["event-1"] = "Title 1"
		c.Category = append(c.Category, Category{Emoji: "🎮", Name: "Game"})
	})
	if existed {
		t.Error("Set() = true, want false for a new channel")
	}
}

func TestSetExistingContent(t *testing.T) {
	d := New("")
	d.Set("channel-1", func(c *Content) {
		c.Events["event-1"] = "Title 1"
	})

	existed := d.Set("channel-1", func(c *Content) {
		if c.Events["event-1"] != "Title 1" {
			t.Errorf(`Events["event-1"] = %q, want %q`, c.Events["event-1"], "Title 1")
		}
		c.Events["event-2"] = "Title 2"
	})
	if !existed {
		t.Error("Set() = false, want true for an existing channel")
	}
}

func TestGetExistingChannel(t *testing.T) {
	d := New("")
	d.Set("channel-1", func(c *Content) {
		c.Events["event-1"] = "Title 1"
		c.Category = append(c.Category, Category{Emoji: "🎮", Name: "Game"})
	})

	called := false
	ok := d.Get("channel-1", func(c *Content) {
		called = true
		if c.Events["event-1"] != "Title 1" {
			t.Errorf(`Events["event-1"] = %q, want %q`, c.Events["event-1"], "Title 1")
		}
		if len(c.Category) != 1 || c.Category[0] != (Category{Emoji: "🎮", Name: "Game"}) {
			t.Errorf("Category = %+v, want [{🎮 Game}]", c.Category)
		}
	})
	if !ok {
		t.Error("Get() = false, want true for an existing channel")
	}
	if !called {
		t.Error("callback was not invoked for an existing channel")
	}
}

func TestGetMissingChannel(t *testing.T) {
	d := New("")
	ok := d.Get("no-such-channel", func(c *Content) {
		t.Error("callback was invoked for a missing channel")
	})
	if ok {
		t.Error("Get() = true, want false for a missing channel")
	}
}

func TestLoadCreatesFileWhenMissing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.bson")
	d := New(path)
	if err := d.Load(); err != nil {
		t.Fatalf("Load() error = %v, want nil for a missing file", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("Load() did not create the file: %v", err)
	}
}

func TestLoadInvalidBSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.bson")
	if err := os.WriteFile(path, []byte("not bson"), 0644); err != nil {
		t.Fatal(err)
	}

	d := New(path)
	if err := d.Load(); err == nil {
		t.Fatal("Load() error = nil, want error for invalid BSON")
	}
}

func TestSyncLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.bson")

	d1 := New(path)
	d1.Set("channel-1", func(c *Content) {
		c.Events["event-1"] = "Title 1"
		c.Category = append(c.Category, Category{Emoji: "🎮", Name: "Game"})
	})
	d1.Set("channel-2", func(c *Content) {
		c.Events["event-2"] = "Title 2"
	})
	if err := d1.Sync(); err != nil {
		t.Fatalf("Sync() error = %v", err)
	}

	d2 := New(path)
	if err := d2.Load(); err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if ok := d2.Get("channel-1", func(c *Content) {
		if c.Events["event-1"] != "Title 1" {
			t.Errorf(`channel-1 Events["event-1"] = %q, want %q`, c.Events["event-1"], "Title 1")
		}
		if len(c.Category) != 1 || c.Category[0] != (Category{Emoji: "🎮", Name: "Game"}) {
			t.Errorf("channel-1 Category = %+v, want [{🎮 Game}]", c.Category)
		}
	}); !ok {
		t.Error("channel-1 not found after round trip")
	}
	if ok := d2.Get("channel-2", func(c *Content) {
		if c.Events["event-2"] != "Title 2" {
			t.Errorf(`channel-2 Events["event-2"] = %q, want %q`, c.Events["event-2"], "Title 2")
		}
	}); !ok {
		t.Error("channel-2 not found after round trip")
	}
}

func TestConcurrentSetGet(t *testing.T) {
	d := New("")
	var wg sync.WaitGroup
	for i := range 100 {
		wg.Add(2)
		go func() {
			defer wg.Done()
			d.Set("channel", func(c *Content) {
				c.Events[strconv.Itoa(i)] = "title"
			})
		}()
		go func() {
			defer wg.Done()
			d.Get("channel", func(c *Content) {
				_ = len(c.Events)
			})
		}()
	}
	wg.Wait()

	if ok := d.Get("channel", func(c *Content) {
		if len(c.Events) != 100 {
			t.Errorf("len(Events) = %d, want 100", len(c.Events))
		}
	}); !ok {
		t.Error("Get() = false after concurrent Set")
	}
}
