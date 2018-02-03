package main

const minBuffer = 10

func createResource(resource string, args []interface{}) error {
	ts, err := newTaxiiStorer()
	if err != nil {
		return err
	}
	defer ts.disconnect()

	toWrite := make(chan interface{}, minBuffer)
	errs := make(chan error, minBuffer)

	go ts.create(resource, toWrite, errs)
	toWrite <- args
	close(toWrite)

	for e := range errs {
		err = e
	}

	return err
}

func readResource(resource string, args []interface{}) (interface{}, error) {
	var result interface{}

	ts, err := newTaxiiStorer()
	if err != nil {
		return result, err
	}
	defer ts.disconnect()

	return ts.read(resource, args)
}
