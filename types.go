package pass

// Credentials is a single set of credentials
type Credentials struct {
	Username *string
	Password *string
}

// Item contains path data and credentials for a single identity
type Item struct {
	Path        []*string
	Credentials *Credentials
}
