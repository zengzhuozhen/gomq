package broker

import "testing"

func TestRunMemberBroker(t *testing.T){
	opts := NewOption(Member, "127.0.0.1:9001", []string{"127.0.0.1:2379"})
	NewBroker(opts).Run()
}