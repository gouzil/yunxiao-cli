package keyring

import "github.com/zalando/go-keyring"

type Backend struct{}

func (Backend) Get(service string, user string) (string, error) {
	return keyring.Get(service, user)
}

func (Backend) Set(service string, user string, password string) error {
	return keyring.Set(service, user, password)
}

func (Backend) Delete(service string, user string) error {
	return keyring.Delete(service, user)
}
