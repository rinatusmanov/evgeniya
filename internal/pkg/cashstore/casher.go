package cashstore

// Cache интерфейс для хранения в кэше данных
// здесь и сейчас данные хранятся в собственной памяти
// однако в дальнейшем можно перевести хранение например в редиску.
type Cache interface {
	Cache(string)
	Check(string) bool
}

type cache struct {
	data map[string]struct{}
}

func (c *cache) Cache(s string) {
	c.data[s] = struct{}{}
}

func (c *cache) Check(s string) bool {
	_, ok := c.data[s]

	return ok
}

//nolint:nolintlint //nolint:ireturn
func NewCache() Cache {
	return &cache{
		data: make(map[string]struct{}),
	}
}
