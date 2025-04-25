package subscribers

type Service struct {
	brokerClient brokerClient
}

func New(bc brokerClient) *Service {
	return &Service{
		brokerClient: bc,
	}
}
