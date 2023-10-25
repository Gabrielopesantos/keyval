package item

import "time"

type Item struct {
	// Item identifier
	// Cannot contain spaces, escaped characters
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

// RETURN Pointer or actual item?
func New(key string, value []byte, flags uint8, ttl uint64) *Item {
	now := time.Now()
	return &Item{
		Key:       key,
		Value:     value,
		Flags:     flags,
		TTL:       ttl,
		createdAt: now,
		updatedAt: now,
		isValid:   true,
	}
}
