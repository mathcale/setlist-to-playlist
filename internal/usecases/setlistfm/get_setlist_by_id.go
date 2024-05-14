package setlistfm

import (
	setlistfm_client "github.com/mathcale/setlist-to-playlist/internal/clients/setlistfm"
	"github.com/mathcale/setlist-to-playlist/internal/entities/setlistfm"
)

type GetSetlistByIDUseCaseInterface interface {
	Execute(id string) (*setlistfm.Set, error)
}

type GetSetlistByIDUseCase struct {
	SetlistFMClient setlistfm_client.SetlistFMClientInterface
}

func NewGetSetlistByIDUseCase(c setlistfm_client.SetlistFMClientInterface) GetSetlistByIDUseCaseInterface {
	return &GetSetlistByIDUseCase{
		SetlistFMClient: c,
	}
}

func (u *GetSetlistByIDUseCase) Execute(id string) (*setlistfm.Set, error) {
	return u.SetlistFMClient.GetSetlistByID(id)
}
