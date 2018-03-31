package main

type routableCollections struct {
	CollectionIDs []taxiiID
}

func (rc *routableCollections) read(ts taxiiStorer, rootPath string) error {
	routables := *rc

	result, err := ts.read("routableCollections", []interface{}{rootPath})
	if err != nil {
		return err
	}

	routables = result.(routableCollections)
	*rc = routables
	return err
}
