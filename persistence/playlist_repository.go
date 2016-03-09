package persistence

import (
	"errors"

	"github.com/deluan/gosonic/domain"
)

type playlistRepository struct {
	ledisRepository
}

func NewPlaylistRepository() domain.PlaylistRepository {
	r := &playlistRepository{}
	r.init("playlist", &domain.Playlist{})
	return r
}

func (r *playlistRepository) Put(m *domain.Playlist) error {
	if m.Id == "" {
		return errors.New("Playlist Id is not set")
	}
	return r.saveOrUpdate(m.Id, m)
}

func (r *playlistRepository) Get(id string) (*domain.Playlist, error) {
	var rec interface{}
	rec, err := r.readEntity(id)
	return rec.(*domain.Playlist), err
}

func (r *playlistRepository) GetAll(options domain.QueryOptions) (*domain.Playlists, error) {
	var as = make(domain.Playlists, 0)
	err := r.loadAll(&as, options)
	return &as, err
}

func (r *playlistRepository) PurgeInactive(active *domain.Playlists) error {
	currentIds, err := r.getAllIds()
	if err != nil {
		return err
	}
	for _, a := range *active {
		currentIds[a.Id] = false
	}
	inactiveIds := make(map[string]bool)
	for id, inactive := range currentIds {
		if inactive {
			inactiveIds[id] = true
		}
	}
	return r.DeleteAll(inactiveIds)
}

var _ domain.PlaylistRepository = (*playlistRepository)(nil)