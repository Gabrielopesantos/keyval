package item

import "time"

type Item struct {
	// Item identifier (Cannot contain spaces or escaped characters)
	Key string
	// Item content
	Value []byte
	// Optional item user defined flags
	Flags uint8
	// Item time to live in seconds
	TTL       uint64
	createdAt time.Time
	updatedAt time.Time
	isValid   bool
}

func New(key string, value []byte, flags uint8, ttl uint64) *Item {
	item := InitItem()
	item.Key = key
	item.Value = value
	item.Flags = flags
	item.TTL = ttl

	return item
}

func InitItem() *Item {
	now := time.Now()
	return &Item{
		createdAt: now,
		updatedAt: now,
		isValid:   true,
	}
}

func (i *Item) Touch() {
	i.updatedAt = time.Now()
}
