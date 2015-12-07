package test

type fakeLog struct{}

func (f *fakeLog) SetLevel(int)                   {}
func (f *fakeLog) Debug(string, ...interface{})   {}
func (f *fakeLog) Info(string, ...interface{})    {}
func (f *fakeLog) Warning(string, ...interface{}) {}
func (f *fakeLog) Warn(string, ...interface{})    {}
func (f *fakeLog) Error(string, ...interface{})   {}
func (f *fakeLog) Fatal(string, ...interface{})   {}
