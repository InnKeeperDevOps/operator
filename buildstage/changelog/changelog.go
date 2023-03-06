package changelog

type Changes struct {
	index map[string]*[]string
}

func New() Changes {
	var c Changes
	if c.index == nil {
		c.index = map[string]*[]string{}
	}
	return c
}

func (c Changes) Add(key string, entry string) {
	if c.index[key] == nil {
		c.index[key] = &[]string{}
	}
	changeArray := append(*c.index[key], entry)
	c.index[key] = &changeArray
}

func (c Changes) Count() int {
	return len(c.index)
}
func (c Changes) GetKeys() []string {
	var keys []string
	for k, _ := range c.index {
		keys = append(keys, k)
	}
	return keys
}
func (c Changes) GetList(key string) *[]string {
	list := c.index[key]
	if list != nil {
		return list
	}
	return nil
}
func (c Changes) Get() map[string]*[]string {
	return c.index
}
