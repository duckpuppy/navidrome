package subsonic

import (
	"context"
	"errors"
	"net/http"

	"github.com/deluan/navidrome/engine"
	"github.com/deluan/navidrome/log"
	"github.com/deluan/navidrome/server/subsonic/responses"
	"github.com/deluan/navidrome/utils"
)

type AlbumListController struct {
	listGen       engine.ListGenerator
	listFunctions map[string]strategy
}

func NewAlbumListController(listGen engine.ListGenerator) *AlbumListController {
	c := &AlbumListController{
		listGen: listGen,
	}
	c.listFunctions = map[string]strategy{
		"random":               c.listGen.GetRandom,
		"newest":               c.listGen.GetNewest,
		"recent":               c.listGen.GetRecent,
		"frequent":             c.listGen.GetFrequent,
		"highest":              c.listGen.GetHighest,
		"alphabeticalByName":   c.listGen.GetByName,
		"alphabeticalByArtist": c.listGen.GetByArtist,
		"starred":              c.listGen.GetStarred,
	}
	return c
}

type strategy func(ctx context.Context, offset int, size int) (engine.Entries, error)

func (c *AlbumListController) getAlbumList(r *http.Request) (engine.Entries, error) {
	typ, err := RequiredParamString(r, "type", "Required string parameter 'type' is not present")
	if err != nil {
		return nil, err
	}
	listFunc, found := c.listFunctions[typ]

	if !found {
		log.Error(r, "albumList type not implemented", "type", typ)
		return nil, errors.New("Not implemented!")
	}

	offset := utils.ParamInt(r, "offset", 0)
	size := utils.MinInt(utils.ParamInt(r, "size", 10), 500)

	albums, err := listFunc(r.Context(), offset, size)
	if err != nil {
		log.Error(r, "Error retrieving albums", "error", err)
		return nil, errors.New("Internal Error")
	}

	return albums, nil
}

func (c *AlbumListController) GetAlbumList(w http.ResponseWriter, r *http.Request) (*responses.Subsonic, error) {
	albums, err := c.getAlbumList(r)
	if err != nil {
		return nil, NewError(responses.ErrorGeneric, err.Error())
	}

	response := NewResponse()
	response.AlbumList = &responses.AlbumList{Album: ToChildren(albums)}
	return response, nil
}

func (c *AlbumListController) GetAlbumList2(w http.ResponseWriter, r *http.Request) (*responses.Subsonic, error) {
	albums, err := c.getAlbumList(r)
	if err != nil {
		return nil, NewError(responses.ErrorGeneric, err.Error())
	}

	response := NewResponse()
	response.AlbumList2 = &responses.AlbumList{Album: ToAlbums(albums)}
	return response, nil
}

func (c *AlbumListController) GetStarred(w http.ResponseWriter, r *http.Request) (*responses.Subsonic, error) {
	artists, albums, mediaFiles, err := c.listGen.GetAllStarred(r.Context())
	if err != nil {
		log.Error(r, "Error retrieving starred media", "error", err)
		return nil, NewError(responses.ErrorGeneric, "Internal Error")
	}

	response := NewResponse()
	response.Starred = &responses.Starred{}
	response.Starred.Artist = ToArtists(artists)
	response.Starred.Album = ToChildren(albums)
	response.Starred.Song = ToChildren(mediaFiles)
	return response, nil
}

func (c *AlbumListController) GetStarred2(w http.ResponseWriter, r *http.Request) (*responses.Subsonic, error) {
	artists, albums, mediaFiles, err := c.listGen.GetAllStarred(r.Context())
	if err != nil {
		log.Error(r, "Error retrieving starred media", "error", err)
		return nil, NewError(responses.ErrorGeneric, "Internal Error")
	}

	response := NewResponse()
	response.Starred2 = &responses.Starred{}
	response.Starred2.Artist = ToArtists(artists)
	response.Starred2.Album = ToAlbums(albums)
	response.Starred2.Song = ToChildren(mediaFiles)
	return response, nil
}

func (c *AlbumListController) GetNowPlaying(w http.ResponseWriter, r *http.Request) (*responses.Subsonic, error) {
	npInfos, err := c.listGen.GetNowPlaying(r.Context())
	if err != nil {
		log.Error(r, "Error retrieving now playing list", "error", err)
		return nil, NewError(responses.ErrorGeneric, "Internal Error")
	}

	response := NewResponse()
	response.NowPlaying = &responses.NowPlaying{}
	response.NowPlaying.Entry = make([]responses.NowPlayingEntry, len(npInfos))
	for i, entry := range npInfos {
		response.NowPlaying.Entry[i].Child = ToChild(entry)
		response.NowPlaying.Entry[i].UserName = entry.UserName
		response.NowPlaying.Entry[i].MinutesAgo = entry.MinutesAgo
		response.NowPlaying.Entry[i].PlayerId = entry.PlayerId
		response.NowPlaying.Entry[i].PlayerName = entry.PlayerName
	}
	return response, nil
}

func (c *AlbumListController) GetRandomSongs(w http.ResponseWriter, r *http.Request) (*responses.Subsonic, error) {
	size := utils.MinInt(utils.ParamInt(r, "size", 10), 500)
	genre := utils.ParamString(r, "genre")

	songs, err := c.listGen.GetRandomSongs(r.Context(), size, genre)
	if err != nil {
		log.Error(r, "Error retrieving random songs", "error", err)
		return nil, NewError(responses.ErrorGeneric, "Internal Error")
	}

	response := NewResponse()
	response.RandomSongs = &responses.Songs{}
	response.RandomSongs.Songs = make([]responses.Child, len(songs))
	for i, entry := range songs {
		response.RandomSongs.Songs[i] = ToChild(entry)
	}
	return response, nil
}
