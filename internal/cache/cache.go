package cache

import (
	"ecksbee.com/telefacts/pkg/cache"
	gocache "github.com/patrickmn/go-cache"
)

func NewCache() *gocache.Cache {
	return cache.NewCache(false)
}

func MarshalRenderable(id string, hash string) ([]byte, error) {
	return cache.MarshalRenderable(id, hash)
}

func MarshalCatalog(id string) ([]byte, error) {
	return cache.MarshalCatalog(id)
}
