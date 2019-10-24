package configrepo

func (cr *ConfigRepo) pushError(err error) {
	if err != nil {
		cr.errorsCh <- err
	}
}

func (cr *ConfigRepo) AddErrorListener(fn ErrorHandlerFn) {
	cr.errorHandlers = append(cr.errorHandlers, fn)
}

func (cr *ConfigRepo) errorHandlerManager() {
	go func() {
		for {
			select {
			case err := <-cr.errorsCh:
				for _, errFn := range cr.errorHandlers {
					errFn(err)
				}
			}
		}
	}()
}
