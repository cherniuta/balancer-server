package core

type Balancer interface {
	NextIndex() int
	GetNextPeer() *Backend
}
