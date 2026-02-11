package handlers

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/h2non/filetype"
	"github.com/vanjmali/spotlite/common-lib/pagination"
	"github.com/vanjmali/spotlite/common-lib/requests"
	"github.com/vanjmali/spotlite/common-lib/respond"
	"github.com/vanjmali/spotlite/common-lib/telemetry"
	"github.com/vanjmali/spotlite/content/dtos"
	"github.com/vanjmali/spotlite/content/services"
)

// SongHandler wires HTTP handlers to the song service and validators.
type SongHandler struct {
	s *services.SongService
	v *validator.Validate
}

// NewSongHandler creates and returns a new SongHandler with the provided service and validator.
func NewSongHandler(s services.SongService, v validator.Validate) *SongHandler {
	h := SongHandler{s: &s, v: &v}
	return &h
}

// HandleCreateSong handles HTTP POST requests to create a new song.
func (h *SongHandler) HandleCreateSong(w http.ResponseWriter, r *http.Request) {
	var req dtos.SongDto

	if ok, err := requests.ReadAndValidateJson(w, h.v, r.Body, &req); !ok {
		if err != nil {
			log.Printf("trace_id=%s invalid request body: %v", telemetry.TraceID(r.Context()), err)
			_ = respond.BadRequest(w, "invalid request body")
		}
		return
	}

	if _, err := h.s.Create(r.Context(), &req); err != nil {
		switch {
		case errors.Is(err, services.ErrObjectIdCastFailed):
			_ = respond.BadRequest(w, "Invalid ID format")
			return
		case errors.Is(err, services.ErrArtistNotFound):
			_ = respond.NotFound(w)
			return
		case errors.Is(err, services.ErrAlbumNotFound):
			_ = respond.NotFound(w)
			return
		case errors.Is(err, services.ErrGenreNotFound):
			_ = respond.NotFound(w)
			return
		default:
			log.Printf("trace_id=%s failed to create song: %v", telemetry.TraceID(r.Context()), err)
			_ = respond.InternalServerError(w)
			return
		}
	}

	respond.NoContent(w)
}

// HandleGetSongById handles HTTP GET requests to retrieve a single song by its ID.
func (h *SongHandler) HandleGetSongById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	song, err := h.s.FindSongById(r.Context(), id)
	if err != nil {
		log.Printf("trace_id=%s failed to get song: %v", telemetry.TraceID(r.Context()), err)

		switch {
		case errors.Is(err, services.ErrSongNotFound):
			_ = respond.NotFound(w)
			return
		case errors.Is(err, services.ErrObjectIdCastFailed):
			_ = respond.BadRequest(w, "Invalid ID format")
			return
		default:
			_ = respond.InternalServerError(w)
			return
		}
	}

	if err := respond.OkJson(w, song); err != nil {
		log.Printf("trace_id=%s failed to write get song response: %v", telemetry.TraceID(r.Context()), err)
	}
}

// HandleUpdateSong handles HTTP PATCH requests to update an existing song.
func (h *SongHandler) HandleUpdateSong(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var dto dtos.UpdateSongDto
	if ok, err := requests.ReadAndValidateJson(w, h.v, r.Body, &dto); !ok {
		if err != nil {
			log.Printf("trace_id=%s invalid request body: %v", telemetry.TraceID(r.Context()), err)
			_ = respond.BadRequest(w, "invalid request body")
		}
		return
	}

	updatedSong, err := h.s.UpdateSong(r.Context(), id, dto)
	switch {
	case errors.Is(err, services.ErrObjectIdCastFailed):
		log.Printf("trace_id=%s invalid song id: %v", telemetry.TraceID(r.Context()), err)
		_ = respond.BadRequest(w, "invalid song id")
		return
	case errors.Is(err, services.ErrSongNotFound):
		log.Printf("trace_id=%s song not found: %v", telemetry.TraceID(r.Context()), err)
		_ = respond.NotFound(w)
		return
	case errors.Is(err, services.ErrArtistNotFound):
		log.Printf("trace_id=%s artist not found: %v", telemetry.TraceID(r.Context()), err)
		_ = respond.NotFound(w)
		return
	case err != nil:
		log.Printf("trace_id=%s failed to update song: %v", telemetry.TraceID(r.Context()), err)
		_ = respond.InternalServerError(w)
		return
	}

	if err := respond.OkJson(w, updatedSong); err != nil {
		log.Printf("trace_id=%s failed to write update song response: %v", telemetry.TraceID(r.Context()), err)
		_ = respond.InternalServerError(w)
		return
	}
}

// HandleDeleteSong handles HTTP DELETE requests to delete a song.
func (h *SongHandler) HandleDeleteSong(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	err := h.s.DeleteSong(r.Context(), id)
	switch {
	case errors.Is(err, services.ErrObjectIdCastFailed):
		log.Printf("trace_id=%s invalid song id: %v", telemetry.TraceID(r.Context()), err)
		_ = respond.BadRequest(w, "invalid song id")
		return
	case errors.Is(err, services.ErrSongNotFound):
		log.Printf("trace_id=%s song not found: %v", telemetry.TraceID(r.Context()), err)
		_ = respond.NotFound(w)
		return
	case err != nil:
		log.Printf("trace_id=%s failed to delete song: %v", telemetry.TraceID(r.Context()), err)
		_ = respond.InternalServerError(w)
		return
	}

	respond.NoContent(w)
}

// HandleGetSongs handles HTTP GET requests to retrieve a paginated list of songs with optional filtering.
func (h *SongHandler) HandleGetSongs(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	p := pagination.ParsePagination(q)
	query := services.SongsQuery{
		Page:     p.Page,
		Size:     p.Size,
		Title:    q.Get("title"),
		Genre:    q.Get("genre"),
		GenreID:  q.Get("genre_id"),
		ArtistId: q.Get("artist_id"),
	}

	handleListResponse(w, r, "songs", func(ctx context.Context) (any, error) {
		return h.s.GetSongs(ctx, query)
	})
}

func (h *SongHandler) HandleUploadSongAudio(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 100<<20)

	if err := r.ParseMultipartForm(100 << 20); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			_ = respond.PayloadTooLarge(w, "file too large (max 100MB)")
			return
		}
		_ = respond.BadRequest(w, "invalid multipart form")
		return
	}

	if r.MultipartForm != nil {
		defer r.MultipartForm.RemoveAll()
	}

	mf := r.MultipartForm
	if mf == nil || mf.File == nil {
		_ = respond.BadRequest(w, "missing file")
		return
	}

	totalFiles := 0
	for _, list := range mf.File {
		totalFiles += len(list)
	}

	if totalFiles != 1 {
		_ = respond.BadRequest(w, "request must contain exactly one file")
		return
	}

	if len(mf.File["file"]) != 1 {
		_ = respond.BadRequest(w, "exactly one file must be provided under field 'file'")
		return
	}

	id := mux.Vars(r)["id"]

	file, header, err := r.FormFile("file")
	if err != nil {
		_ = respond.BadRequest(w, "missing file")
		return
	}
	defer file.Close()

	log.Printf("trace_id=%s uploaded file: name=%q, size=%d",
		telemetry.TraceID(r.Context()), header.Filename, header.Size)

	head := make([]byte, 512)
	n, err := file.Read(head)
	if err != nil && err != io.EOF {
		log.Printf("trace_id=%s failed to read file header: %v", telemetry.TraceID(r.Context()), err)
		_ = respond.BadRequest(w, "failed to read file")
		return
	}
	if n == 0 {
		_ = respond.BadRequest(w, "empty file")
		return
	}

	sniff := head[:n]

	kind, err := filetype.Match(sniff)
	if err != nil {
		log.Printf("trace_id=%s unable to determine file type: %v", telemetry.TraceID(r.Context()), err)
		_ = respond.BadRequest(w, "unable to determine file type")
		return
	}
	if kind == filetype.Unknown {
		log.Printf("trace_id=%s unknown file type (filename=%q)", telemetry.TraceID(r.Context()), header.Filename)
		_ = respond.BadRequest(w, "unknown file type")
		return
	}

	allowedMimes := map[string]string{
		"audio/mpeg":   ".mp3",
		"audio/wav":    ".wav",
		"audio/x-wav":  ".wav",
		"audio/flac":   ".flac",
		"audio/x-flac": ".flac",
		"audio/aac":    ".aac",
		"audio/ogg":    ".ogg",
		"audio/opus":   ".opus",
		"audio/mp4":    ".m4a",
	}

	mime := kind.MIME.Value
	ext, ok := allowedMimes[mime]
	if !ok {
		log.Printf("trace_id=%s invalid file type: detected_extension=%s, detected_mime=%s, filename=%q",
			telemetry.TraceID(r.Context()), kind.Extension, mime, header.Filename)
		_ = respond.BadRequest(w, "file is not a valid audio file")
		return
	}

	log.Printf("trace_id=%s file type validated: extension=%s, mime=%s",
		telemetry.TraceID(r.Context()), ext, mime)

	reader := io.MultiReader(bytes.NewReader(sniff), file)

	updated, err := h.s.UploadAudio(r.Context(), id, reader, ext, mime)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrSongNotFound):
			log.Printf("trace_id=%s song not found: %s", telemetry.TraceID(r.Context()), id)
			_ = respond.NotFound(w)
			return
		case errors.Is(err, services.ErrAudioUploadFailed):
			log.Printf("trace_id=%s audio upload failed for song: %s, error: %v", telemetry.TraceID(r.Context()), id, err)
			_ = respond.InternalServerError(w)
			return
		case errors.Is(err, services.ErrObjectIdCastFailed):
			log.Printf("trace_id=%s invalid id format: %s", telemetry.TraceID(r.Context()), id)
			_ = respond.BadRequest(w, "Invalid ID format")
			return
		default:
			log.Printf("trace_id=%s unexpected error during audio upload for song %s: %v", telemetry.TraceID(r.Context()), id, err)
			_ = respond.InternalServerError(w)
			return
		}
	}

	_ = respond.OkJson(w, updated)
}

func (h *SongHandler) HandleStreamSongAudio(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	song, err := h.s.FindSongById(r.Context(), id)
	if err != nil {
		log.Printf("trace_id=%s failed to get song: %v", telemetry.TraceID(r.Context()), err)

		switch {
		case errors.Is(err, services.ErrSongNotFound):
			_ = respond.NotFound(w)
			return
		case errors.Is(err, services.ErrObjectIdCastFailed):
			_ = respond.BadRequest(w, "Invalid ID format")
			return
		default:
			_ = respond.InternalServerError(w)
			return
		}
	}

	if song.AudioPath == "" {
		_ = respond.NotFound(w)
		return
	}

	rc, err := h.s.OpenAudio(r.Context(), song.AudioPath)
	if err != nil {
		log.Printf("trace_id=%s failed to open audio file at path %s: %v", telemetry.TraceID(r.Context()), song.AudioPath, err)
		_ = respond.InternalServerError(w)
		return
	}
	defer rc.Close()

	if song.AudioSize > 0 {
		w.Header().Set("Content-Length", strconv.FormatInt(song.AudioSize, 10))
	}
	if song.AudioMimeType != "" {
		w.Header().Set("Content-Type", song.AudioMimeType)
	} else {
		w.Header().Set("Content-Type", "audio/mpeg")
	}

	_, _ = io.Copy(w, rc)
}

func (h *SongHandler) HandleCreateSongWithAudio(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 100<<20)

	if err := r.ParseMultipartForm(100 << 20); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			_ = respond.PayloadTooLarge(w, "payload too large (max 100MB)")
			return
		}
		_ = respond.BadRequest(w, "invalid multipart form")
		return
	}
	if r.MultipartForm != nil {
		defer r.MultipartForm.RemoveAll()
	}

	metaStr := r.FormValue("meta")
	if metaStr == "" {
		_ = respond.BadRequest(w, "missing meta")
		return
	}

	var dto dtos.SongDto
	if ok, err := requests.ReadAndValidateJson(w, h.v, io.NopCloser(bytes.NewReader([]byte(metaStr))), &dto); !ok {
		if err != nil {
			log.Printf(
				"trace_id=%s invalid meta payload: %v", telemetry.TraceID(r.Context()), err)
		}
		_ = respond.BadRequest(w, "invalid meta")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		_ = respond.BadRequest(w, "missing file")
		return
	}
	defer file.Close()

	head := make([]byte, 512)
	n, err := file.Read(head)
	if err != nil && err != io.EOF {
		_ = respond.BadRequest(w, "failed to read file")
		return
	}
	if n == 0 {
		_ = respond.BadRequest(w, "empty file")
		return
	}
	sniff := head[:n]

	kind, err := filetype.Match(sniff)
	if err != nil || kind == filetype.Unknown {
		_ = respond.BadRequest(w, "unknown file type")
		return
	}

	allowedMimes := map[string]string{
		"audio/mpeg":   ".mp3",
		"audio/wav":    ".wav",
		"audio/x-wav":  ".wav",
		"audio/flac":   ".flac",
		"audio/x-flac": ".flac",
		"audio/aac":    ".aac",
		"audio/ogg":    ".ogg",
		"audio/opus":   ".opus",
		"audio/mp4":    ".m4a",
	}

	mime := kind.MIME.Value
	ext, ok := allowedMimes[mime]
	if !ok {
		log.Printf("trace_id=%s invalid file type: detected_extension=%s, detected_mime=%s, filename=%q",
			telemetry.TraceID(r.Context()), kind.Extension, mime, header.Filename)
		_ = respond.BadRequest(w, "file is not a valid audio file")
		return
	}

	reader := io.MultiReader(bytes.NewReader(sniff), file)

	id, err := h.s.Create(r.Context(), &dto)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrObjectIdCastFailed):
			_ = respond.BadRequest(w, "invalid id format")
		case errors.Is(err, services.ErrArtistNotFound), errors.Is(err, services.ErrGenreNotFound):
			_ = respond.NotFound(w)
		default:
			_ = respond.InternalServerError(w)
		}
		return
	}

	updated, err := h.s.UploadAudio(r.Context(), id.Hex(), reader, ext, mime)
	if err != nil {
		_ = h.s.DeleteSong(r.Context(), id.Hex()) // rollback

		switch {
		case errors.Is(err, services.ErrAudioUploadFailed):
			_ = respond.InternalServerError(w)
		default:
			_ = respond.InternalServerError(w)
		}
		return
	}

	_ = respond.OkJson(w, updated)
}
