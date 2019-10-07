package storage

// Code generated by cdproto-gen. DO NOT EDIT.

// EventCacheStorageContentUpdated a cache's contents have been modified.
//
// See: https://chromedevtools.github.io/devtools-protocol/tot/Storage#event-cacheStorageContentUpdated
type EventCacheStorageContentUpdated struct {
	Origin    string `json:"origin"`    // Origin to update.
	CacheName string `json:"cacheName"` // Name of cache in origin.
}

// EventCacheStorageListUpdated a cache has been added/deleted.
//
// See: https://chromedevtools.github.io/devtools-protocol/tot/Storage#event-cacheStorageListUpdated
type EventCacheStorageListUpdated struct {
	Origin string `json:"origin"` // Origin to update.
}

// EventIndexedDBContentUpdated the origin's IndexedDB object store has been
// modified.
//
// See: https://chromedevtools.github.io/devtools-protocol/tot/Storage#event-indexedDBContentUpdated
type EventIndexedDBContentUpdated struct {
	Origin          string `json:"origin"`          // Origin to update.
	DatabaseName    string `json:"databaseName"`    // Database to update.
	ObjectStoreName string `json:"objectStoreName"` // ObjectStore to update.
}

// EventIndexedDBListUpdated the origin's IndexedDB database list has been
// modified.
//
// See: https://chromedevtools.github.io/devtools-protocol/tot/Storage#event-indexedDBListUpdated
type EventIndexedDBListUpdated struct {
	Origin string `json:"origin"` // Origin to update.
}