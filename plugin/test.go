package plugin

type TestPlugin struct {
	Status string
	Info   string
}

func NewTestPlugin() *TestPlugin {
	return &TestPlugin{}
}

func (plugin* TestPlugin) DoStuff() int {
	return 12
}
