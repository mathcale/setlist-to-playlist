package setlistfm

import (
	setlistfm_client "github.com/mathcale/setlist-to-playlist/internal/clients/setlistfm"
	"github.com/mathcale/setlist-to-playlist/internal/entities/setlistfm"
)

type GetSetlistByIDUseCaseInterface interface {
	Execute(input setlistfm.GetSetlistByIDInput) (*setlistfm.Set, error)
}

type GetSetlistByIDUseCase struct {
	SetlistFMClient setlistfm_client.SetlistFMClientInterface
}

func NewGetSetlistByIDUseCase(c setlistfm_client.SetlistFMClientInterface) GetSetlistByIDUseCaseInterface {
	return &GetSetlistByIDUseCase{
		SetlistFMClient: c,
	}
}

func (u *GetSetlistByIDUseCase) Execute(in setlistfm.GetSetlistByIDInput) (*setlistfm.Set, error) {
	id, err := in.SetlistID()
	if err != nil {
		return nil, err
	}

	return u.SetlistFMClient.GetSetlistByID(*id)
}
