package broker

import "testing"

func TestRunLeaderBroker(t *testing.T){
	opts := NewOption(Leader, "0.0.0.0:9000", []string{"0.0.0.0:2379"})
	NewBroker(opts).Run()
}
