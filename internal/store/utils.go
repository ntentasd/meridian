package store

func GetKey(kind, namespace, name string) string {
	return kind + "/" + namespace + "/" + name
}
