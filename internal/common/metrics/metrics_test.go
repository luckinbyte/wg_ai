package metrics

import "testing"

func TestMetrics(t *testing.T) {
	m := NewMetrics()
	m.IncConnections()
	m.IncMessages()
	m.IncErrors()

	if m.GetConnections() != 1 {
		t.Error("connections should be 1")
	}
	if m.GetMessages() != 1 {
		t.Error("messages should be 1")
	}
	if m.GetErrors() != 1 {
		t.Error("errors should be 1")
	}
}

func TestMetricsDecrement(t *testing.T) {
	m := NewMetrics()
	m.IncConnections()
	m.IncConnections()
	m.DecConnections()

	if m.GetConnections() != 1 {
		t.Error("connections should be 1 after increment twice and decrement once")
	}
}

func TestMetricsConcurrent(t *testing.T) {
	m := NewMetrics()
	done := make(chan bool)

	for i := 0; i < 100; i++ {
		go func() {
			m.IncConnections()
			m.IncMessages()
			done <- true
		}()
	}

	for i := 0; i < 100; i++ {
		<-done
	}

	if m.GetConnections() != 100 {
		t.Errorf("connections should be 100, got %d", m.GetConnections())
	}
	if m.GetMessages() != 100 {
		t.Errorf("messages should be 100, got %d", m.GetMessages())
	}
}
