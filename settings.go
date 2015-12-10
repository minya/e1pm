package main

type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type PushoverSettings struct {
	Token string `json:"token"`
	User  string `json:"user"`
}

type Settings struct {
	Credentials Credentials      `json:"credentials"`
	Pushover    PushoverSettings `json:"pushover"`
}
