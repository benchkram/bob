package cache

type Cache interface {
	Get(key string) (string, bool)
	Save(key, value string) (err error)
}
