package main

type Credentials struct {
	email    string
	password string
}

type Settings struct {
	credentials Credentials
}
