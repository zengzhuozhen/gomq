package broker

import "testing"

func TestRunLeaderBroker(t *testing.T) {
	opts := NewOption(Leader, "127.0.0.1:9000", "/var/log/tempmq.log", true, []string{"127.0.0.1:2379"})
	NewBroker(opts).Run()
}
