package broker

import "testing"

func TestRunMemberBroker(t *testing.T){
	NewBroker(ServerType(Member),
		EndPoint("127.0.0.1:9001"),
		Dirname("/var/log/tempmq.log"),
		EtcdUrl([]string{"127.0.0.1:2379"}),
	).Run()
}