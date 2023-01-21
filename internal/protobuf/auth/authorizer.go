package auth

import (
	"fmt"

	"github.com/casbin/casbin/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Authorizer struct {
	enforcer *casbin.Enforcer
}

func New(model, plicy string) (*Authorizer, error) {
	enforcer, err := casbin.NewEnforcer(model, plicy)
	return &Authorizer{enforcer: enforcer}, err
}

func (a *Authorizer) Authorize(subject, object, action string) error {
	ok, err := a.enforcer.Enforce(subject, object, action)
	if err != nil {
		return err
	}
	if !ok {
		msg := fmt.Sprintf("%s not permitted to %s to %s", subject, action, object)
		st := status.New(codes.PermissionDenied, msg)
		return st.Err()
	}
	return nil
}
