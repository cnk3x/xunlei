package utils

func SeqExec(fns ...func() (err error)) (err error) {
	for _, fn := range fns {
		if err = fn(); err != nil {
			break
		}
	}
	return
}
