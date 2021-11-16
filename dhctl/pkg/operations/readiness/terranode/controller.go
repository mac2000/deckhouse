package terranode

type ReadinessChecker struct{}

func NewChecker() *ReadinessChecker {
	return &ReadinessChecker{}
}

func (c *ReadinessChecker) IsNodeGroupReady(_ ...string) (bool, error) {
	// simple terranode node group always ready
	return true, nil
}
