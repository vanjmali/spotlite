package handlers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
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

const maxSongAudioUploadBytes = 100 << 20
const maxSongAudioUploadMessage = "payload too large (max 100MB)"

var allowedSongAudioMimes = map[string]string{
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

var (
	errMultipartTooLarge    = errors.New("multipart payload too large")
	errInvalidMultipartForm = errors.New("invalid multipart form")
)

// NewSongHandler creates and returns a new SongHandler with the provided service and validator.
func NewSongHandler(s services.SongService, v validator.Validate) *SongHandler {
	h := SongHandler{s: &s, v: &v}
	return &h
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
	if err := parseMultipartWithLimit(w, r); err != nil {
		switch {
		case errors.Is(err, errMultipartTooLarge):
			_ = respond.PayloadTooLarge(w, maxSongAudioUploadMessage)
		default:
			_ = respond.BadRequest(w, "invalid multipart form")
		}
		return
	}
	defer r.MultipartForm.RemoveAll()

	id := mux.Vars(r)["id"]

	file, reader, ext, mime, ok := getSingleValidatedAudioUpload(w, r)
	if !ok {
		return
	}
	defer file.Close()

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
	if err := parseMultipartWithLimit(w, r); err != nil {
		switch {
		case errors.Is(err, errMultipartTooLarge):
			_ = respond.PayloadTooLarge(w, maxSongAudioUploadMessage)
		default:
			_ = respond.BadRequest(w, "invalid multipart form")
		}
		return
	}
	defer r.MultipartForm.RemoveAll()

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

	file, reader, ext, mime, ok := getSingleValidatedAudioUpload(w, r)
	if !ok {
		return
	}
	defer file.Close()

	id, err := h.s.Create(r.Context(), &dto)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrObjectIdCastFailed):
			_ = respond.BadRequest(w, "invalid id format")
		case errors.Is(err, services.ErrArtistNotFound), errors.Is(err, services.ErrGenreNotFound), errors.Is(err, services.ErrAlbumNotFound):
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

func parseMultipartWithLimit(w http.ResponseWriter, r *http.Request) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxSongAudioUploadBytes)

	if err := r.ParseMultipartForm(maxSongAudioUploadBytes); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			return errMultipartTooLarge
		}
		return fmt.Errorf("%w: %v", errInvalidMultipartForm, err)
	}

	return nil
}

// getSingleValidatedAudioUpload enforces one file input, verifies audio type from magic bytes,
// and rebuilds a full reader (sniffed bytes + remaining stream) for downstream upload.
func getSingleValidatedAudioUpload(w http.ResponseWriter, r *http.Request) (multipart.File, io.Reader, string, string, bool) {
	mf := r.MultipartForm
	if mf == nil || mf.File == nil {
		_ = respond.BadRequest(w, "missing file")
		return nil, nil, "", "", false
	}

	totalFiles := 0
	for _, list := range mf.File {
		totalFiles += len(list)
	}
	if totalFiles != 1 {
		_ = respond.BadRequest(w, "request must contain exactly one file")
		return nil, nil, "", "", false
	}
	if len(mf.File["file"]) != 1 {
		_ = respond.BadRequest(w, "exactly one file must be provided under field 'file'")
		return nil, nil, "", "", false
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		_ = respond.BadRequest(w, "missing file")
		return nil, nil, "", "", false
	}

	sniff, err := readFileSniff(file)
	if err != nil {
		_ = respond.BadRequest(w, err.Error())
		_ = file.Close()
		return nil, nil, "", "", false
	}

	ext, mime, badRequestMessage, ok := validateAudioFileSniff(r.Context(), sniff, header.Filename)
	if !ok {
		_ = respond.BadRequest(w, badRequestMessage)
		_ = file.Close()
		return nil, nil, "", "", false
	}

	log.Printf("trace_id=%s uploaded file: name=%q, size=%d, mime=%s, extension=%s",
		telemetry.TraceID(r.Context()), header.Filename, header.Size, mime, ext)

	// file.Read already consumed sniff bytes; prepend them so upload reads the entire original file.
	reader := io.MultiReader(bytes.NewReader(sniff), file)
	return file, reader, ext, mime, true
}

// readFileSniff reads up to 512 header bytes used for content-type detection by magic bytes.
func readFileSniff(file multipart.File) ([]byte, error) {
	head := make([]byte, 512)
	n, err := file.Read(head)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, errors.New("failed to read file")
	}
	if n == 0 {
		return nil, errors.New("empty file")
	}

	return head[:n], nil
}

// validateAudioFileSniff identifies the file type from header bytes and allows only configured audio MIME types.
func validateAudioFileSniff(ctx context.Context, sniff []byte, filename string) (string, string, string, bool) {
	kind, err := filetype.Match(sniff)
	if err != nil {
		log.Printf("trace_id=%s unable to determine file type: %v", telemetry.TraceID(ctx), err)
		return "", "", "unable to determine file type", false
	}
	if kind == filetype.Unknown {
		log.Printf("trace_id=%s unknown file type (filename=%q)", telemetry.TraceID(ctx), filename)
		return "", "", "unknown file type", false
	}

	mime := kind.MIME.Value
	ext, ok := allowedSongAudioMimes[mime]
	if !ok {
		log.Printf("trace_id=%s invalid file type: detected_extension=%s, detected_mime=%s, filename=%q",
			telemetry.TraceID(ctx), kind.Extension, mime, filename)
		return "", "", "file is not a valid audio file", false
	}

	return ext, mime, "", true
}
