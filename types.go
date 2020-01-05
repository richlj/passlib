package pass

// Credentials is a single set of credentials
type Credentials struct {
	Username string
	Password string
}

// Item contains path data and credentials for a single identity
type Item struct {
	Path        Path
	Credentials Credentials
}

// Items contains data about zero or more Items
type Items struct {
	Items []Item
}

type password struct {
	value string
}

// Path is the address of an item, formed of elements
type Path struct {
	Elements []element
}

type element struct {
	Element string
}

type directory struct {
	Directory string
}

type itemWrapper struct {
	value Item
}
