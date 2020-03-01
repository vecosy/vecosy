package gitconfigrepo

func (cr *GitConfigRepo) pushError(err error) {
	if err != nil {
		cr.errorsCh <- err
	}
}

// AddErrorListener add an error handler to the git repo
func (cr *GitConfigRepo) AddErrorListener(fn ErrorHandlerFn) {
	cr.errorHandlers = append(cr.errorHandlers, fn)
}

func (cr *GitConfigRepo) errorHandlerManager() {
	go func() {
		for {
			err := <-cr.errorsCh
			for _, errFn := range cr.errorHandlers {
				errFn(err)
			}
		}
	}()
}
