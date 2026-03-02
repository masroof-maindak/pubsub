package utils

type ClientType int

const (
	Subscriber ClientType = iota
	Publisher
)

var ConnType = map[ClientType]string{
	Subscriber: "subscriber",
	Publisher:  "publisher",
}
