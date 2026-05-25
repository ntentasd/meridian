package store

import (
	"time"

	"k8s.io/apimachinery/pkg/types"
)

type RouteEntry struct {
	UID       types.UID
	Name      string
	Namespace string
	Kind      string
	Hostnames []string
	Rules     []RouteRule

	// annotation-enriched
	Owner       string
	Description string
	DocsURL     string
	Tags        []string

	// housekeeping
	UpdatedAt time.Time
	Labels    map[string]string
}

type RouteRule struct {
	Path    string
	Methods []string
}

type EventKind string

const (
	EventUpsert EventKind = "upsert"
	EventDelete EventKind = "delete"
)

type Event struct {
	Kind  EventKind
	Entry RouteEntry
}
