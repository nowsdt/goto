package web

import (
	. "goto/store"
)

var store Store

func GetStore() *Store {
	return &store
}
func SetStore(s *Store) {
	store = *s
}
