package configGitRepo

func (cr *GitConfigRepo) pushError(err error) {
	if err != nil {
		cr.errorsCh <- err
	}
}

func (cr *GitConfigRepo) AddErrorListener(fn ErrorHandlerFn) {
	cr.errorHandlers = append(cr.errorHandlers, fn)
}

func (cr *GitConfigRepo) errorHandlerManager() {
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
