package songs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	database "fucku/internal/database"
	utils "fucku/internal/utils"
)

type Song struct {
	Id              string        `json:"id"`
	SongName        string        `json:"song_name"`
	AlbumName       string        `json:"album_name"`
	Artist          string        `json:"artist"`
	FeaturingArtist string        `json:"featuring_artist"`
	SpotifyEmbedUrl string        `json:"spotify_embed_url"`
	Reason          string        `json:"reason"`
	Genres          *[]Genre      `json:"genres"`
	SpecialTags     *[]SpecialTag `json:"special_tags"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
}

func (s *Song) validateSongName(errors map[string]string) {
	s.SongName = strings.TrimSpace(s.SongName)
	if len(s.SongName) == 0 {
		errors["song_name"] = "no song name provided"
	}
}

func (s *Song) validateAlbumName(errors map[string]string) {
	s.AlbumName = strings.TrimSpace(s.AlbumName)
	if len(s.AlbumName) == 0 {
		errors["album_name"] = "no album name provided"
	}
}

func (s *Song) validateArtistName(errors map[string]string) {
	s.Artist = strings.TrimSpace(s.Artist)
	if len(s.Artist) == 0 {
		errors["aritst"] = "no artist name provided"
	}
}

func (s *Song) validateSpotifyEmbedUrl(errors map[string]string) {
	s.SpotifyEmbedUrl = strings.TrimSpace(s.SpotifyEmbedUrl)
	if len(s.SpotifyEmbedUrl) == 0 {
		errors["spotify_embed"] = "no spotify embed url provided"
	}

	re := regexp.MustCompile(`src="([^"]+)"`)
	matches := re.FindStringSubmatch(s.SpotifyEmbedUrl)

	if len(matches) > 1 {
		s.SpotifyEmbedUrl = matches[1]
	} else {
		errors["spotify_embed_2"] = "no source found"
	}
}

func (s *Song) validateReason(errors map[string]string) {
	s.Reason = strings.TrimSpace(s.Reason)
	if len(s.Reason) == 0 {
		errors["reason"] = "no reason provided"
	}
}

type Genre struct {
	Id        int       `json:"id"`
	GenreName string    `json:"genre_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (g *Genre) validateGenreName(errors map[string]string) {
	g.GenreName = strings.TrimSpace(g.GenreName)
	if len(g.GenreName) == 0 {
		errors["genre_name"] = "no genre name provided"
	}
}

type SpecialTag struct {
	Id          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (st *SpecialTag) validateName(errors map[string]string) {
	st.Name = strings.TrimSpace(st.Name)
	if len(st.Name) == 0 {
		errors["name"] = "no tag name provided"
	}
}

func GetSongs(db *database.Database, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Pagination: default to page 1, 25 items per page
		page := 1
		if p := r.URL.Query().Get("page"); p != "" {
			if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
				page = parsed
			}
		}
		limit := 25
		offset := (page - 1) * limit

		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		// First, fetch songs with pagination
		rows, err := db.DBPool.Query(ctx, `
			SELECT id, song_name, album_name, artist, featuring_artist, spotify_embed_url, reason, created_at, updated_at
			FROM songs
			ORDER BY created_at DESC
			LIMIT $1 OFFSET $2
		`, limit, offset)
		if err != nil {
			logger.Error("failed to fetch songs", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var songs []Song
		for rows.Next() {
			var s Song
			if err := rows.Scan(
				&s.Id, &s.SongName, &s.AlbumName, &s.Artist, &s.FeaturingArtist,
				&s.SpotifyEmbedUrl, &s.Reason, &s.CreatedAt, &s.UpdatedAt,
			); err != nil {
				logger.Error("failed to scan song row", "error", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			songs = append(songs, s)
		}
		if err := rows.Err(); err != nil {
			logger.Error("row iteration error", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if len(songs) == 0 {
			json.NewEncoder(w).Encode([]Song{})
			return
		}

		// Build a slice of IDs for fetching genres and tags
		songIDs := make([]string, len(songs))
		songIndex := make(map[string]int)
		for i, s := range songs {
			songIDs[i] = s.Id
			songIndex[s.Id] = i
		}

		// Fetch genres
		genreRows, err := db.DBPool.Query(ctx, `
			SELECT sg.song_id, g.id, g.genre_name, g.created_at, g.updated_at
			FROM song_genres sg
			JOIN genres g ON sg.genre_id = g.id
			WHERE sg.song_id = ANY($1)
		`, songIDs)
		if err != nil {
			logger.Error("failed to fetch genres", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		defer genreRows.Close()

		for genreRows.Next() {
			var songID string
			var g Genre
			if err := genreRows.Scan(&songID, &g.Id, &g.GenreName, &g.CreatedAt, &g.UpdatedAt); err != nil {
				logger.Error("failed to scan genre row", "error", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			idx := songIndex[songID]
			if songs[idx].Genres == nil {
				songs[idx].Genres = &[]Genre{}
			}
			*songs[idx].Genres = append(*songs[idx].Genres, g)
		}

		// Fetch special tags
		tagRows, err := db.DBPool.Query(ctx, `
			SELECT st.song_id, t.id, t.name, t.description, t.created_at, t.updated_at
			FROM song_special_tags st
			JOIN special_tags t ON st.tag_id = t.id
			WHERE st.song_id = ANY($1)
		`, songIDs)
		if err != nil {
			logger.Error("failed to fetch special tags", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		defer tagRows.Close()

		for tagRows.Next() {
			var songID string
			var t SpecialTag
			if err := tagRows.Scan(&songID, &t.Id, &t.Name, &t.Description, &t.CreatedAt, &t.UpdatedAt); err != nil {
				logger.Error("failed to scan special tag row", "error", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			idx := songIndex[songID]
			if songs[idx].SpecialTags == nil {
				songs[idx].SpecialTags = &[]SpecialTag{}
			}
			*songs[idx].SpecialTags = append(*songs[idx].SpecialTags, t)
		}

		// Return as JSON
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(songs)
	})
}

func CreateSong(db *database.Database, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var s Song
		err := utils.DecodeJSONBody(w, r, &s)
		if err != nil {
			var mr *utils.MalformedRequest
			if errors.As(err, &mr) {
				http.Error(w, mr.Msg, mr.Status)
				return
			} else {
				logger.Error("error while decoding json body in create song", "error", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}

		// Validation
		errs := make(map[string]string)

		s.validateSongName(errs)
		s.validateAlbumName(errs)
		s.validateArtistName(errs)
		s.validateReason(errs)
		s.validateSpotifyEmbedUrl(errs)

		if len(errs) != 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]any{
				"errors": errs,
			})

			return
		}

		// Insertion
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		var id string
		row := db.DBPool.QueryRow(ctx, `
			INSERT INTO songs (
			song_name,
			album_name,
			artist,
			featuring_artist,
			spotify_embed_url,
			reason
			) VALUES ($1, $2, $3, $4, $5, $6) 
			RETURNING id
			`, s.SongName, s.AlbumName, s.Artist, s.FeaturingArtist, s.SpotifyEmbedUrl, s.Reason)

		if err := row.Scan(&id); err != nil {
			logger.Error("error while trying to insert song into database", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		logger.Info("song created", "song", s, "id", id)

		w.WriteHeader(201)
		fmt.Fprint(w, id)
	})
}

func GetGenres(db *database.Database, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		rows, err := db.DBPool.Query(ctx, `
			SELECT id, genre_name, created_at, updated_at
			FROM genres
			ORDER BY genre_name ASC 
		`)
		if err != nil {
			logger.Error("failed to fetch genres", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var genres []Genre
		for rows.Next() {
			var g Genre
			if err := rows.Scan(&g.Id, &g.GenreName, &g.CreatedAt, &g.UpdatedAt); err != nil {
				logger.Error("failed to scan genre row", "error", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			genres = append(genres, g)
		}
		if err := rows.Err(); err != nil {
			logger.Error("row iteration error", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if len(genres) == 0 {
			json.NewEncoder(w).Encode([]Genre{})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(genres)
	})
}

func CreateGenre(db *database.Database, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var genre Genre
		err := utils.DecodeJSONBody(w, r, &genre)
		if err != nil {
			var mr *utils.MalformedRequest
			if errors.As(err, &mr) {
				http.Error(w, mr.Msg, mr.Status)
				return
			} else {
				logger.Error("error while decoding json body in create song", "error", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}

		// Validation
		errs := make(map[string]string)

		genre.validateGenreName(errs)
		if len(errs) != 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]any{
				"errors": errs,
			})
			return
		}

		// Insertion
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		var id int
		row := db.DBPool.QueryRow(ctx, `
			INSERT INTO genres (
			genre_name
			) VALUES ($1) 
			RETURNING id
			`, genre.GenreName)

		if err := row.Scan(&id); err != nil {
			logger.Error("error while trying to insert song into database", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		logger.Info("genre created", "genre", genre, "id", id)

		w.WriteHeader(201)
		fmt.Fprint(w, id)
	})
}

func CreateSpecialTag(db *database.Database, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var tag SpecialTag
		err := utils.DecodeJSONBody(w, r, &tag)
		if err != nil {
			var mr *utils.MalformedRequest
			if errors.As(err, &mr) {
				http.Error(w, mr.Msg, mr.Status)
				return
			} else {
				logger.Error("error while decoding json body in create song", "error", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}

		// Validation
		errs := make(map[string]string)

		tag.validateName(errs)
		if len(errs) != 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]any{
				"errors": errs,
			})
			return
		}

		// Insertion
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		var id int
		row := db.DBPool.QueryRow(ctx, `
			INSERT INTO special_tags (
            name,
            description
			) VALUES ($1, $2) 
			RETURNING id
			`, tag.Name, tag.Description)

		if err := row.Scan(&id); err != nil {
			logger.Error("error while trying to insert song into database", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		logger.Info("tag created", "tag", tag, "id", id)

		w.WriteHeader(201)
		fmt.Fprint(w, id)
	})
}

func GetSpecialTags(db *database.Database, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		rows, err := db.DBPool.Query(ctx, `
			SELECT id, name, description, created_at, updated_at
			FROM special_tags
			ORDER BY name ASC 
		`)
		if err != nil {
			logger.Error("failed to fetch tags", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var tags []SpecialTag
		for rows.Next() {
			var tag SpecialTag
			if err := rows.Scan(&tag.Id, &tag.Name, &tag.Description, &tag.CreatedAt, &tag.UpdatedAt); err != nil {
				logger.Error("failed to scan genre row", "error", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			tags = append(tags, tag)
		}
		if err := rows.Err(); err != nil {
			logger.Error("row iteration error", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if len(tags) == 0 {
			json.NewEncoder(w).Encode([]Genre{})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(tags)
	})
}

func AssignGenre(db *database.Database, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		songId := r.URL.Query().Get("song")
		genreId := r.URL.Query().Get("genre")

		if len(songId) == 0 || len(genreId) == 0 {
			http.Error(w, "required query parameters (song, genre) not found", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		_, err := db.DBPool.Exec(ctx, `
            INSERT INTO song_genres (song_id, genre_id) VALUES ($1, $2)
            `, songId, genreId)
		if err != nil {
			logger.Error("failed to assign genre to song", "song", songId, "genre", genreId, "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(200)
	})
}

func AssignTag(db *database.Database, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		songId := r.URL.Query().Get("song")
		tagId := r.URL.Query().Get("tag")

		if len(songId) == 0 || len(tagId) == 0 {
			http.Error(w, "required query parameters (song, tag) not found", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		_, err := db.DBPool.Exec(ctx, `
            INSERT INTO song_special_tags (song_id, tag_id) VALUES ($1, $2)
            `, songId, tagId)
		if err != nil {
			logger.Error("failed to assign tag to song", "song", songId, "tag", tagId, "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(200)
	})
}

func UnassignGenre(db *database.Database, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		songId := r.URL.Query().Get("song")
		genreId := r.URL.Query().Get("genre")

		if len(songId) == 0 || len(genreId) == 0 {
			http.Error(w, "required query parameters (song, genre) not found", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		_, err := db.DBPool.Exec(ctx, `
            DELETE FROM song_genres WHERE song_id = $1 AND genre_id = $2
            `, songId, genreId)
		if err != nil {
			logger.Error("failed to remove genre from song", "song", songId, "genre", genreId, "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(200)
	})
}

func UnassignTag(db *database.Database, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		songId := r.URL.Query().Get("song")
		tagId := r.URL.Query().Get("tag")

		if len(songId) == 0 || len(tagId) == 0 {
			http.Error(w, "required query parameters (song, tag) not found", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		_, err := db.DBPool.Exec(ctx, `
            DELETE FROM song_special_tags WHERE song_id = $1 AND tag_id = $2)
            `, songId, tagId)
		if err != nil {
			logger.Error("failed to assign tag to song", "song", songId, "tag", tagId, "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(200)
	})
}
