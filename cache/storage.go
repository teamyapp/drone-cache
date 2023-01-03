package cache

type Storage interface {
	PersistCache() error
	RetrieveCache() error
}
