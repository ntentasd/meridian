package store

import (
	"time"

	"k8s.io/apimachinery/pkg/types"
)

type RouteEntry struct {
	UID       types.UID   `json:"uid"`
	Name      string      `json:"name"`
	Namespace string      `json:"namespace"`
	Kind      string      `json:"kind"`
	Hostname  string      `json:"hostname"`
	Rules     []RouteRule `json:"rules"`

	// annotation-enriched
	LogoURL     string   `json:"logoUrl"`
	Owner       string   `json:"owner"`
	Description string   `json:"description"`
	DocsURL     string   `json:"docsUrl"`
	Tags        []string `json:"tags"`

	// housekeeping
	UpdatedAt time.Time         `json:"updatedAt"`
	Labels    map[string]string `json:"labels"`
}

type RouteRule struct {
	Path    string   `json:"path"`
	Methods []string `json:"methods"`
}

type EventKind string

const (
	EventUpsert EventKind = "upsert"
	EventDelete EventKind = "delete"
)

type Event struct {
	Kind  EventKind  `json:"kind"`
	Entry RouteEntry `json:"route"`
}
