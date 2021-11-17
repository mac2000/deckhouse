package terraform

type InfraActionHook interface {
	BeforeAction() (runAfterAction bool, err error)
	IsReady() error
	AfterAction() error
}

type DummyHook struct{}

func (c *DummyHook) BeforeAction() (runPostAction bool, err error) {
	return false, nil
}

func (c *DummyHook) IsReady() error {
	return nil
}

func (c *DummyHook) AfterAction() error {
	return nil
}
