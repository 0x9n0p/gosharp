package immreturn

func Caller() error {
	Callee()?
}

func Callee() error {
	return nil
}
