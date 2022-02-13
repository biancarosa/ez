package ez

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}
