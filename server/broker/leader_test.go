package broker

import "testing"

func TestRunLeaderBroker(t *testing.T) {
	NewBroker(ServerType(Leader),
		EndPoint("127.0.0.1:9000"),
		Dirname("/var/log/tempmq.log"),
		EtcdUrl([]string{"127.0.0.1:2379"}),
	).Run()
}
