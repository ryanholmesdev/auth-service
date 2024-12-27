package services

import (
	"golang.org/x/oauth2"
	"sync"
)

type TokenStore struct {
	store sync.Map
}

var tokenStore = TokenStore{}

func StoreToken(sessionID string, provider string, token *oauth2.Token) {
	tokenStore.store.Store(sessionID+"_"+provider, token)
}

func GetToken(sessionID string, provider string) (*oauth2.Token, bool) {
	value, ok := tokenStore.store.Load(sessionID + "_" + provider)
	if !ok {
		return nil, false
	}
	return value.(*oauth2.Token), true
}
