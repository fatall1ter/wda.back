// Package domain contains domain data models and business rules
//
// Date: 2020-11-24
package domain

// User properties
type User struct {
	UserID       int64  `json:"user_id,omitempty"`
	UserName     string `json:"user_name,omitempty"`
	UserFullName string `json:"user_full_name,omitempty"`
	DomainName   string `json:"domain_name,omitempty"`
	Login        string `json:"login,omitempty"`
	PWord        string `json:"pword,omitempty"`
	Post         string `json:"post,omitempty"`
	EMail        string `json:"email,omitempty"`
	Telefon      string `json:"phone,omitempty"`
	SMTP         string `json:"smtp,omitempty"`
	EMailPWord   string `json:"email_password,omitempty"`
	Options      string `json:"options,omitempty"`
	Comment      string `json:"comment,omitempty"`
}

// Users slice of users
type Users []User

// UserRepoI behavior of user repo
type UserRepoI interface {
	AddUser(User) (*User, error)
	GetUsers(int64, int64) (Users, int64, error)
	GetUserByID(int64) (*User, error)
	UserSetPass(int64, string) error
	DelUser(int64) error
	Login(uLogin, uPass string) (*User, error)
	GetSrvPortDB() string
	HealthCheck() error
}
