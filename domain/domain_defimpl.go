package domain

// Code generated by defimpl for defaul implenebtation of interfaces. DO NOT EDIT.

// DefImplUserRepoI default implementation of UserRepoI
type DefImplUserRepoI struct{}

// AddUser default implementation method of UserRepoI interface
func (DefImplUserRepoI) AddUser(User) (*User, error) {
	panic("method AddUser not implemented")
}

// GetUsers default implementation method of UserRepoI interface
func (DefImplUserRepoI) GetUsers(int64, int64) (Users, int64, error) {
	panic("method GetUsers not implemented")
}

// GetUserByID default implementation method of UserRepoI interface
func (DefImplUserRepoI) GetUserByID(int64) (*User, error) {
	panic("method GetUserByID not implemented")
}

// UserSetPass default implementation method of UserRepoI interface
func (DefImplUserRepoI) UserSetPass(int64, string) error {
	panic("method UserSetPass not implemented")
}

// DelUser default implementation method of UserRepoI interface
func (DefImplUserRepoI) DelUser(int64) error {
	panic("method DelUser not implemented")
}

// Login default implementation method of UserRepoI interface
func (DefImplUserRepoI) Login(string, string) (*User, error) {
	panic("method Login not implemented")
}

// GetSrvPortDB default implementation method of UserRepoI interface
func (DefImplUserRepoI) GetSrvPortDB() string {
	panic("method GetSrvPortDB not implemented")
}

// HealthCheck default implementation method of UserRepoI interface
func (DefImplUserRepoI) HealthCheck() error {
	panic("method HealthCheck not implemented")
}