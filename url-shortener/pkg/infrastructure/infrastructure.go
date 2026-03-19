// Package infrastructure facilitates interaction with repository and external services
package infrastructure

import (
	"github.com/cmmasaba/prototypes/urlshortener/pkg/infrastructure/repository"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/infrastructure/services/hibp"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/infrastructure/services/mail"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/infrastructure/services/otp"
)

type Infrastructure struct {
	*repository.Repository
	*hibp.HIBP
	*mail.MailerImpl
	*otp.Provider
}

func New(
	database *repository.Repository,
	hibp *hibp.HIBP,
	mail *mail.MailerImpl,
	otp *otp.Provider,
) (*Infrastructure, error) {
	return &Infrastructure{
		database,
		hibp,
		mail,
		otp,
	}, nil
}
