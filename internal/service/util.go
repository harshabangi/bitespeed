package service

type Void struct{}

var VoidValue Void

func keyExists(k string, m map[string]Void) bool {
	_, exists := m[k]
	return exists
}
