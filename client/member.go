package client

type Member struct {
	opts   *Option
	client *client
}

func NewMember(opts *Option) *Member {
	client := NewClient(opts).(*client)
	return &Member{client: client,opts:opts}
}




