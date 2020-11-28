package broker

import "testing"

func TestRunMemberBroker(t *testing.T){
	opts := NewOption(Member, "0.0.0.0:9001", []string{"0.0.0.0:2379"})
	NewBroker(opts).Run()
}