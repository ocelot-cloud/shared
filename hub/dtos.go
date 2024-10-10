package hub

type TagInfo struct {
	User string `json:"user"`
	App  string `json:"app"`
	Tag  string `json:"tag"`
}

type AppAndTag struct {
	App string `json:"app"`
	Tag string `json:"tag"`
}

type TagUpload struct {
	App     string `json:"app"`
	Tag     string `json:"tag"`
	Content []byte `json:"content"`
}

type UserAndApp struct {
	User string `json:"user"`
	App  string `json:"app"`
}

type RegistrationForm struct {
	User     string `json:"user"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type LoginCredentials struct {
	User     string `json:"user"`
	Password string `json:"password"`
	// TODO Couldn't I take the origin directly from the request? Seems unnecessary.
	Origin string `json:"origin"`
}
