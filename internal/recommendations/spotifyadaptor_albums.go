package recommendations

import (
	"context"
	"fmt"
	"sort"

	"github.com/kristofferostlund/recommendli/pkg/ctxhelper"
	"github.com/kristofferostlund/recommendli/pkg/spotifypaginator"
	"github.com/zmb3/spotify"
	"golang.org/x/sync/errgroup"
)

func (s *SpotifyAdaptor) GetAlbum(ctx context.Context, albumID string) (spotify.FullAlbum, error) {
	if err := ctxhelper.Closed(ctx); err != nil {
		return spotify.FullAlbum{}, fmt.Errorf("getting album %s: %w", albumID, err)
	}
	return s.getAlbum(ctx, albumID)
}

func (s *SpotifyAdaptor) GetAlbums(ctx context.Context, albumIDs []string) ([]spotify.FullAlbum, error) {
	if err := ctxhelper.Closed(ctx); err != nil {
		return nil, fmt.Errorf("listing albums: %w", err)
	}
	return s.getStoredAlbums(ctx, albumIDs)
}

func (s *SpotifyAdaptor) getAlbum(ctx context.Context, albumID string) (spotify.FullAlbum, error) {
	albums, err := s.getStoredAlbums(ctx, []string{albumID})
	if err != nil {
		return spotify.FullAlbum{}, err
	}
	return albums[0], nil
}

func (s *SpotifyAdaptor) getStoredAlbums(ctx context.Context, albumIDs []string) ([]spotify.FullAlbum, error) {
	stored := make(map[string]spotify.FullAlbum)

	needsFetching := make([]string, 0)
	for _, id := range albumIDs {
		storeKey := fmt.Sprintf("album_%s", id)
		var album spotify.FullAlbum
		if exists, err := s.kv.Get(ctx, storeKey, &album); err == nil && exists {
			stored[id] = album
			continue
		} else if err != nil {
			return nil, fmt.Errorf("getting album %s from store: %w", storeKey, err)
		}

		needsFetching = append(needsFetching, id)
	}

	fetched := make(map[string]spotify.FullAlbum)
	if len(needsFetching) > 0 {
		albums, err := s.getAlbums(ctx, needsFetching)
		if err != nil {
			return nil, fmt.Errorf("getting albums: %w", err)
		}
		for _, album := range albums {
			fetched[album.ID.String()] = album
			storeKey := fmt.Sprintf("album_%s", album.ID)
			if err := s.kv.Put(ctx, storeKey, album); err != nil {
				return nil, fmt.Errorf("updating album store (%s): %w", storeKey, err)
			}
		}
	}

	albums := make([]spotify.FullAlbum, 0)
	for _, id := range albumIDs {
		if album, exists := stored[id]; exists {
			albums = append(albums, album)
			continue
		}
		if album, exists := fetched[id]; exists {
			albums = append(albums, album)
			continue
		}
		// Should never happen territory...
		return nil, fmt.Errorf("album %s doesn't exist", id)
	}
	return albums, nil
}

func (s *SpotifyAdaptor) getAlbums(ctx context.Context, albumIDs []string) ([]spotify.FullAlbum, error) {
	type indexAndAlbums struct {
		index  int
		albums []spotify.FullAlbum
	}

	paginator := spotifypaginator.New(spotifypaginator.Parallelism(10), spotifypaginator.PageSize(20), spotifypaginator.InitialTotalCount(len(albumIDs)))
	albumChan := make(chan indexAndAlbums)
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		defer close(albumChan)
		return paginator.Run(ctx, func(index int, opts spotifypaginator.PageOpts, next spotifypaginator.NextFunc) (result *spotifypaginator.NextResult, err error) {
			from, to := opts.Offset, opts.Offset+opts.Limit
			spotifyIDs := make([]spotify.ID, 0)
			for _, id := range albumIDs[from:to] {
				spotifyIDs = append(spotifyIDs, spotify.ID(id))
			}

			albumPtrs, err := s.spotify.GetAlbums(spotifyIDs...)
			if err != nil {
				return nil, err
			}
			albums := make([]spotify.FullAlbum, 0)
			for i, a := range albumPtrs {
				if a == nil {
					return nil, fmt.Errorf("album %s doesn't exist", spotifyIDs[i])
				}
				albums = append(albums, *a)
			}
			s.log.Debug("getting albums", "total size", len(albumIDs), "batch size", len(spotifyIDs), "from", from, "to", to)
			albumChan <- indexAndAlbums{index, albums}
			return next(len(albumIDs)), nil
		})
	})

	indexedAlbums := make([]indexAndAlbums, 0)
	for ia := range albumChan {
		indexedAlbums = append(indexedAlbums, ia)
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}

	sort.Slice(indexedAlbums, func(i, j int) bool {
		return indexedAlbums[i].index < indexedAlbums[j].index
	})
	albums := make([]spotify.FullAlbum, 0)
	for _, indexed := range indexedAlbums {
		albums = append(albums, indexed.albums...)
	}
	return albums, nil
}

func (s *SpotifyAdaptor) ListArtistAlbums(ctx context.Context, artistID string) ([]spotify.SimpleAlbum, error) {
	type indexAndAlbums struct {
		index  int
		albums []spotify.SimpleAlbum
	}

	paginator := spotifypaginator.New(spotifypaginator.Parallelism(10))
	albumChan := make(chan indexAndAlbums)
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		defer close(albumChan)
		return paginator.Run(ctx, func(index int, opts spotifypaginator.PageOpts, next spotifypaginator.NextFunc) (result *spotifypaginator.NextResult, err error) {
			page, err := s.spotify.GetArtistAlbumsOpt(spotify.ID(artistID), spotifyOpts(opts), spotify.AlbumTypeAlbum, spotify.AlbumTypeSingle)
			if err != nil {
				return nil, err
			}
			albumChan <- indexAndAlbums{index, page.Albums}
			return next(page.Total), nil
		})
	})

	indexedAlbums := make([]indexAndAlbums, 0)
	for ia := range albumChan {
		indexedAlbums = append(indexedAlbums, ia)
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}

	sort.Slice(indexedAlbums, func(i, j int) bool {
		return indexedAlbums[i].index < indexedAlbums[j].index
	})
	albums := make([]spotify.SimpleAlbum, 0)
	for _, indexed := range indexedAlbums {
		albums = append(albums, indexed.albums...)
	}
	return albums, nil
}
