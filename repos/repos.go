// Package repos contains lowlevel databases details
//
// Date: 2020-11-24
package repos

import (
	"context"
	"errors"
	"time"

	"git.countmax.ru/countmax/cmaxdb"
	"git.countmax.ru/countmax/wda.back/domain"
	"go.uber.org/zap"
)

// ErrLoginPass stat error about check credentials
var ErrLoginPass = errors.New("login or pass didn't match")

// CMRepo implementation of the domain.IUserRepo
type CMRepo struct {
	domain.DefImplUserRepoI
	timeout time.Duration
	cm      *cmaxdb.CMAX
	log     *zap.SugaredLogger
}

// NewCMRepo makes new instance of the CMRepo/domain.IUserRepo
func NewCMRepo(cs string, timeout time.Duration, logger *zap.SugaredLogger) (*CMRepo, error) {
	cm, err := cmaxdb.NewCMAX(cs)
	if err != nil {
		return nil, err
	}
	scope := cm.GetSrvPortDB()
	log := logger.With(zap.String("dbserver", scope))
	cmr := &CMRepo{
		timeout: timeout,
		cm:      cm,
		log:     log,
	}
	return cmr, nil
}

// Login extract user by login and check pass;
// if user not fond error, if pass not match error
func (cmr *CMRepo) Login(uLogin, uPass string) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cmr.timeout)
	defer cancel()
	u, err := cmr.cm.GetUserByLogin(ctx, uLogin)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, ErrLoginPass
	}
	if !cmaxdb.UserСheckPass(u.PWord, uPass) {
		return nil, ErrLoginPass
	}
	du := &domain.User{
		UserID:       u.UserID,
		UserName:     u.UserName,
		UserFullName: u.UserFullName,
		DomainName:   u.DomainName,
		Login:        u.Login,
		PWord:        u.PWord,
		Post:         u.Post,
		EMail:        u.EMail,
		Telefon:      u.Telefon,
		SMTP:         u.SMTP,
		EMailPWord:   u.EMailPWord,
		Options:      u.Options,
		Comment:      u.Comment,
	}
	return du, nil
}

// AddUser creates new user in the repo
func (cmr *CMRepo) AddUser(u domain.User) (*domain.User, error) {
	cu := cmaxdb.User{
		UserID:       u.UserID,
		UserName:     u.UserName,
		UserFullName: u.UserFullName,
		DomainName:   u.DomainName,
		Login:        u.Login,
		PWord:        u.PWord,
		Post:         u.Post,
		EMail:        u.EMail,
		Telefon:      u.Telefon,
		SMTP:         u.SMTP,
		EMailPWord:   u.EMailPWord,
		Options:      u.Options,
		Comment:      u.Comment,
	}
	ctx, cancel := context.WithTimeout(context.Background(), cmr.timeout)
	defer cancel()
	id, err := cmr.cm.CreateUser(ctx, cu)
	if err != nil {
		return nil, err
	}
	u.UserID = id
	return &u, nil
}

// UserSetPass upd password for specified user id
func (cmr *CMRepo) UserSetPass(id int64, pass string) error {
	ctx, cancel := context.WithTimeout(context.Background(), cmr.timeout)
	defer cancel()
	_, err := cmr.cm.UpdUserPass(ctx, id, pass)
	return err
}

// GetUsers extracts users from repo
func (cmr *CMRepo) GetUsers(offset, limit int64) (domain.Users, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cmr.timeout)
	defer cancel()
	cusers, count, err := cmr.cm.GetUsers(ctx, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	if len(cusers) == 0 {
		return nil, 0, nil
	}
	users := make(domain.Users, len(cusers))
	for i, cu := range cusers {
		u := domain.User{
			UserID:       cu.UserID,
			UserName:     cu.UserName,
			UserFullName: cu.UserFullName,
			DomainName:   cu.DomainName,
			Login:        cu.Login,
			Post:         cu.Post,
			EMail:        cu.EMail,
			Telefon:      cu.Telefon,
			SMTP:         cu.SMTP,
			EMailPWord:   cu.EMailPWord,
			Options:      cu.Options,
			Comment:      cu.Comment,
		}
		users[i] = u
	}
	return users, count, nil
}

// GetUserByID extracts specified user from repo
func (cmr *CMRepo) GetUserByID(id int64) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cmr.timeout)
	defer cancel()
	cu, err := cmr.cm.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if cu == nil {
		return nil, nil
	}
	u := &domain.User{
		UserID:       cu.UserID,
		UserName:     cu.UserName,
		UserFullName: cu.UserFullName,
		DomainName:   cu.DomainName,
		Login:        cu.Login,
		Post:         cu.Post,
		EMail:        cu.EMail,
		Telefon:      cu.Telefon,
		SMTP:         cu.SMTP,
		EMailPWord:   cu.EMailPWord,
		Options:      cu.Options,
		Comment:      cu.Comment,
	}

	return u, nil
}

// DelUser erase user record from repo
func (cmr *CMRepo) DelUser(id int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), cmr.timeout)
	defer cancel()
	u := cmaxdb.User{
		UserID: id,
		Login:  "undefined",
	}
	_, err := cmr.cm.DelUser(ctx, u)
	return err
}

// GetSrvPortDB make [server].[port].[db] string
func (cmr *CMRepo) GetSrvPortDB() string {
	return cmr.cm.GetSrvPortDB()
}

// HealthCheck makes ping to database
func (cmr *CMRepo) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), cmr.timeout)
	defer cancel()
	return cmr.cm.HealthWithContext(ctx)
}
