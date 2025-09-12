package resource

type Resource interface { // vheng todo: we may make volume or snapshot implement this
	GetType() string
}
