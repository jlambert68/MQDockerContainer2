package mqcore

import "testing"

func TestGatewayTypeImplementsInterface(t *testing.T) {
	var _ Gateway = (*gateway)(nil)
}

func TestQueueInfoFields(t *testing.T) {
	q := QueueInfo{
		Name:            "DEV.QUEUE.1",
		Description:     "test",
		CurrentDepth:    1,
		MaxDepth:        5000,
		OpenInputCount:  2,
		OpenOutputCount: 3,
	}
	if q.Name == "" || q.MaxDepth == 0 {
		t.Fatalf("unexpected queue info values: %+v", q)
	}
}
