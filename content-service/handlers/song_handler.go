package handlers

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/h2non/filetype"
	"github.com/redis/go-redis/v9"
	commondtos "github.com/vanjmali/spotlite/common-lib/dtos"
	"github.com/vanjmali/spotlite/common-lib/logging"
	"github.com/vanjmali/spotlite/common-lib/pagination"
	"github.com/vanjmali/spotlite/common-lib/requests"
	"github.com/vanjmali/spotlite/common-lib/respond"
	"github.com/vanjmali/spotlite/common-lib/telemetry"
	"github.com/vanjmali/spotlite/content/dtos"
	"github.com/vanjmali/spotlite/content/services"
)

// SongHandler wires HTTP handlers to the song service and validators.
type SongHandler struct {
	s  *services.SongService
	rc *redis.Client
	v  *validator.Validate
}

const (
	maxSongAudioUploadBytes   = 100 << 20
	maxSongAudioUploadMessage = "payload too large (max 100MB)"
)

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

var audioDurationDetector = detectAudioDurationSeconds

// NewSongHandler creates and returns a new SongHandler with the provided service and validator.
func NewSongHandler(s services.SongService, rc *redis.Client, v validator.Validate) *SongHandler {
	h := SongHandler{s: &s, rc: rc, v: &v}
	return &h
}

// HandleCreateSong handles HTTP POST requests to create a new song.
func (h *SongHandler) HandleCreateSong(w http.ResponseWriter, r *http.Request) {
	var req dtos.SongDto

	if ok, err := requests.ReadAndValidateJson(w, h.v, r.Body, &req); !ok {
		if err != nil {
			logging.Warnf(r.Context(), "invalid request body: %v", err)
			_ = respond.BadRequest(w, respond.ErrorMessage("invalid request body"))
		}
		return
	}

	if _, err := h.s.Create(r.Context(), &req); err != nil {
		switch {
		case errors.Is(err, services.ErrObjectIdCastFailed):
			_ = respond.BadRequest(w, respond.ErrorMessage("Invalid ID format"))
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
			logging.Errorf(r.Context(), "failed to create song: %v", err)
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
		logging.Errorf(r.Context(), "failed to get song: %v", err)

		switch {
		case errors.Is(err, services.ErrSongNotFound):
			_ = respond.NotFound(w)
			return
		case errors.Is(err, services.ErrObjectIdCastFailed):
			_ = respond.BadRequest(w, respond.ErrorMessage("Invalid ID format"))
			return
		default:
			_ = respond.InternalServerError(w)
			return
		}
	}

	if err := respond.OkJson(w, song); err != nil {
		logging.Errorf(r.Context(), "failed to write get song response: %v", err)
	}
}

// HandleUpdateSong handles HTTP PATCH requests to update an existing song.
func (h *SongHandler) HandleUpdateSong(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var dto dtos.UpdateSongDto
	if ok, err := requests.ReadAndValidateJson(w, h.v, r.Body, &dto); !ok {
		if err != nil {
			logging.Warnf(r.Context(), "invalid request body: %v", err)
			_ = respond.BadRequest(w, respond.ErrorMessage("invalid request body"))
		}
		return
	}

	updatedSong, err := h.s.UpdateSong(r.Context(), id, dto)
	switch {
	case errors.Is(err, services.ErrObjectIdCastFailed):
		logging.Warnf(r.Context(), "invalid song id: %v", err)
		_ = respond.BadRequest(w, respond.ErrorMessage("invalid song id"))
		return
	case errors.Is(err, services.ErrSongNotFound):
		logging.Warnf(r.Context(), "song not found: %v", err)
		_ = respond.NotFound(w)
		return
	case errors.Is(err, services.ErrArtistNotFound):
		logging.Warnf(r.Context(), "artist not found: %v", err)
		_ = respond.NotFound(w)
		return
	case err != nil:
		logging.Errorf(r.Context(), "failed to update song: %v", err)
		_ = respond.InternalServerError(w)
		return
	}

	if err := respond.OkJson(w, updatedSong); err != nil {
		logging.Errorf(r.Context(), "failed to write update song response: %v", err)
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
		logging.Warnf(r.Context(), "invalid song id: %v", err)
		_ = respond.BadRequest(w, respond.ErrorMessage("invalid song id"))
		return
	case errors.Is(err, services.ErrSongNotFound):
		logging.Warnf(r.Context(), "song not found: %v", err)
		_ = respond.NotFound(w)
		return
	case err != nil:
		logging.Errorf(r.Context(), "failed to delete song: %v", err)
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

	commondtos.HandleListResponse(w, r, "songs", func(ctx context.Context) (any, error) {
		return h.s.GetSongs(ctx, query)
	})
}

func (h *SongHandler) HandleUploadSongAudio(w http.ResponseWriter, r *http.Request) {
	if err := parseMultipartWithLimit(w, r); err != nil {
		switch {
		case errors.Is(err, errMultipartTooLarge):
			logSecurityEvent(r.Context(), "upload_rejected_payload_too_large", "endpoint=song_audio_upload")
			_ = respond.PayloadTooLarge(w, maxSongAudioUploadMessage)
		default:
			logSecurityEvent(r.Context(), "upload_rejected_invalid_multipart", "endpoint=song_audio_upload")
			_ = respond.BadRequest(w, respond.ErrorMessage("invalid multipart form"))
		}
		return
	}
	defer r.MultipartForm.RemoveAll()

	id := mux.Vars(r)["id"]

	file, reader, ext, mime, lengthSeconds, ok := getSingleValidatedAudioUpload(w, r)
	if !ok {
		return
	}
	defer file.Close()

	updated, err := h.s.UploadAudio(r.Context(), id, reader, ext, mime, lengthSeconds)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrSongNotFound):
			logging.Warnf(r.Context(), "song not found: %s", id)
			_ = respond.NotFound(w)
			return
		case errors.Is(err, services.ErrAudioUploadFailed):
			logging.Errorf(r.Context(), "audio upload failed for song: %s, error: %v", id, err)
			_ = respond.InternalServerError(w)
			return
		case errors.Is(err, services.ErrObjectIdCastFailed):
			logging.Warnf(r.Context(), "invalid id format: %s", id)
			_ = respond.BadRequest(w, respond.ErrorMessage("Invalid ID format"))
			return
		default:
			logging.Errorf(r.Context(), "unexpected error during audio upload for song %s: %v", id, err)
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
		logging.Errorf(r.Context(), "failed to get song: %v", err)

		switch {
		case errors.Is(err, services.ErrSongNotFound):
			_ = respond.NotFound(w)
			return
		case errors.Is(err, services.ErrObjectIdCastFailed):
			_ = respond.BadRequest(w, respond.ErrorMessage("Invalid ID format"))
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
		logging.Errorf(r.Context(), "failed to open audio file at path %s: %v", song.AudioPath, err)
		_ = respond.InternalServerError(w)
		return
	}
	if err := verifySongAudioChecksumFromReader(r.Context(), song.AudioPath, song.AudioChecksum, rc); err != nil {
		_ = rc.Close()
		logSecurityEvent(r.Context(), "stream_rejected_integrity_check_failed", fmt.Sprintf("song_id=%s path=%s", id, song.AudioPath))
		_ = respond.InternalServerError(w)
		return
	}
	_ = rc.Close()

	streamReader, err := h.s.OpenAudio(r.Context(), song.AudioPath)
	if err != nil {
		logging.Errorf(r.Context(), "failed to reopen audio file at path %s: %v", song.AudioPath, err)
		_ = respond.InternalServerError(w)
		return
	}
	defer streamReader.Close()

	setSongAudioResponseHeaders(w, song.AudioSize, song.AudioMimeType)

	_, _ = io.Copy(w, streamReader)
}

func (h *SongHandler) HandleCreateSongWithAudio(w http.ResponseWriter, r *http.Request) {
	if err := parseMultipartWithLimit(w, r); err != nil {
		switch {
		case errors.Is(err, errMultipartTooLarge):
			logSecurityEvent(r.Context(), "upload_rejected_payload_too_large", "endpoint=song_create_with_audio")
			_ = respond.PayloadTooLarge(w, maxSongAudioUploadMessage)
		default:
			logSecurityEvent(r.Context(), "upload_rejected_invalid_multipart", "endpoint=song_create_with_audio")
			_ = respond.BadRequest(w, respond.ErrorMessage("invalid multipart form"))
		}
		return
	}
	defer r.MultipartForm.RemoveAll()

	metaStr := r.FormValue("meta")
	if metaStr == "" {
		logSecurityEvent(r.Context(), "upload_rejected_missing_meta", "endpoint=song_create_with_audio")
		_ = respond.BadRequest(w, respond.ErrorMessage("missing meta"))
		return
	}

	var dto dtos.SongDto
	if ok, err := requests.ReadAndValidateJson(w, h.v, io.NopCloser(bytes.NewReader([]byte(metaStr))), &dto); !ok {
		if err != nil {
			logSecurityEvent(r.Context(), "upload_rejected_invalid_meta", fmt.Sprintf("error=%v", err))
		}
		_ = respond.BadRequest(w, respond.ErrorMessage("invalid meta"))
		return
	}

	file, reader, ext, mime, lengthSeconds, ok := getSingleValidatedAudioUpload(w, r)
	if !ok {
		return
	}
	defer file.Close()

	id, err := h.s.Create(r.Context(), &dto)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrObjectIdCastFailed):
			_ = respond.BadRequest(w, respond.ErrorMessage("invalid id format"))
		case errors.Is(err, services.ErrArtistNotFound), errors.Is(err, services.ErrGenreNotFound), errors.Is(err, services.ErrAlbumNotFound):
			_ = respond.NotFound(w)
		default:
			_ = respond.InternalServerError(w)
		}
		return
	}

	updated, err := h.s.UploadAudio(r.Context(), id.Hex(), reader, ext, mime, lengthSeconds)
	if err != nil {
		if rollbackErr := h.s.DeleteSong(r.Context(), id.Hex()); rollbackErr != nil {
			logSecurityEvent(r.Context(), "create_with_audio_rollback_failed", fmt.Sprintf("song_id=%s error=%v", id.Hex(), rollbackErr))
		}

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
		return fmt.Errorf("%w: %w", errInvalidMultipartForm, err)
	}

	return nil
}

// getSingleValidatedAudioUpload enforces one file input, verifies audio type from magic bytes,
// and rebuilds a full reader (sniffed bytes + remaining stream) for downstream upload.
func getSingleValidatedAudioUpload(w http.ResponseWriter, r *http.Request) (multipart.File, io.Reader, string, string, *int, bool) {
	mf := r.MultipartForm
	if mf == nil || mf.File == nil {
		logSecurityEvent(r.Context(), "upload_rejected_missing_file", "reason=no_file_part")
		_ = respond.BadRequest(w, respond.ErrorMessage("missing file"))
		return nil, nil, "", "", nil, false
	}

	totalFiles := 0
	for _, list := range mf.File {
		totalFiles += len(list)
	}
	if totalFiles != 1 {
		logSecurityEvent(r.Context(), "upload_rejected_multiple_files", fmt.Sprintf("count=%d", totalFiles))
		_ = respond.BadRequest(w, respond.ErrorMessage("request must contain exactly one file"))
		return nil, nil, "", "", nil, false
	}
	if len(mf.File["file"]) != 1 {
		logSecurityEvent(r.Context(), "upload_rejected_invalid_file_field", "field=file")
		_ = respond.BadRequest(w, respond.ErrorMessage("exactly one file must be provided under field 'file'"))
		return nil, nil, "", "", nil, false
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		logSecurityEvent(r.Context(), "upload_rejected_missing_file", "reason=form_file_error")
		_ = respond.BadRequest(w, respond.ErrorMessage("missing file"))
		return nil, nil, "", "", nil, false
	}

	sniff, err := readFileSniff(file)
	if err != nil {
		logSecurityEvent(r.Context(), "upload_rejected_invalid_file", "reason="+err.Error())
		_ = respond.BadRequest(w, respond.ErrorMessage(err.Error()))
		_ = file.Close()
		return nil, nil, "", "", nil, false
	}

	ext, mime, badRequestMessage, ok := validateAudioFileSniff(r.Context(), sniff, header.Filename)
	if !ok {
		_ = respond.BadRequest(w, respond.ErrorMessage(badRequestMessage))
		_ = file.Close()
		return nil, nil, "", "", nil, false
	}

	lengthSeconds, err := audioDurationDetector(file, mime)
	if err != nil {
		logSecurityEvent(r.Context(), "upload_rejected_audio_duration_detection_failed", fmt.Sprintf("filename=%q mime=%s", header.Filename, mime))
		_ = respond.BadRequest(w, respond.ErrorMessage(err.Error()))
		_ = file.Close()
		return nil, nil, "", "", nil, false
	}

	if _, err := file.Seek(int64(len(sniff)), io.SeekStart); err != nil {
		logSecurityEvent(r.Context(), "upload_rejected_invalid_file", "reason=failed to reset file cursor")
		_ = respond.BadRequest(w, respond.ErrorMessage("failed to read file"))
		_ = file.Close()
		return nil, nil, "", "", nil, false
	}

	logging.Infof(r.Context(), "uploaded file: name=%q, size=%d, mime=%s, extension=%s",
		header.Filename, header.Size, mime, ext)

	// file.Read already consumed sniff bytes; prepend them so upload reads the entire original file.
	reader := io.MultiReader(bytes.NewReader(sniff), file)
	return file, reader, ext, mime, lengthSeconds, true
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
		logSecurityEvent(ctx, "upload_rejected_type_detection_failed", fmt.Sprintf("filename=%q", filename))
		return "", "", "unable to determine file type", false
	}
	if kind == filetype.Unknown {
		logSecurityEvent(ctx, "upload_rejected_unknown_file_type", fmt.Sprintf("filename=%q", filename))
		return "", "", "unknown file type", false
	}

	mime := kind.MIME.Value
	ext, ok := allowedSongAudioMimes[mime]
	if !ok {
		logSecurityEvent(ctx, "upload_rejected_disallowed_mime", fmt.Sprintf("detected_mime=%s filename=%q", mime, filename))
		return "", "", "file is not a valid audio file", false
	}

	return ext, mime, "", true
}

// detectAudioDurationSeconds derives audio duration from file metadata/header.
// ffprobe is used to support all allowed audio formats.
func detectAudioDurationSeconds(file multipart.File, mime string) (*int, error) {
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, errors.New("failed to read audio duration")
	}
	defer file.Seek(0, io.SeekStart)

	tmp, err := os.CreateTemp("", "song-audio-*.bin")
	if err != nil {
		return nil, errors.New("failed to read audio duration")
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if _, err := io.Copy(tmp, file); err != nil {
		_ = tmp.Close()
		return nil, errors.New("failed to read audio duration")
	}
	if err := tmp.Close(); err != nil {
		return nil, errors.New("failed to read audio duration")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// #nosec G204 - tmpPath is created by os.CreateTemp and not user-controlled.
	cmd := exec.CommandContext(
		ctx,
		"ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		tmpPath,
	)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return nil, errors.New("failed to read audio duration")
	}

	out := strings.TrimSpace(stdout.String())
	if out == "" {
		return nil, errors.New("failed to read audio duration")
	}

	durationSeconds, err := strconv.ParseFloat(out, 64)
	if err != nil || durationSeconds <= 0 {
		return nil, errors.New("failed to read audio duration")
	}

	seconds := max(int(math.Ceil(durationSeconds)), 1)
	return &seconds, nil
}

func verifySongAudioChecksumFromReader(ctx context.Context, audioPath string, expected string, r io.Reader) error {
	if expected == "" {
		return nil
	}

	hasher := sha256.New()
	if _, err := io.Copy(hasher, r); err != nil {
		return fmt.Errorf("failed to read audio for checksum path=%s trace_id=%s", audioPath, telemetry.TraceID(ctx))
	}
	computed := hex.EncodeToString(hasher.Sum(nil))
	if computed == expected {
		return nil
	}

	return fmt.Errorf("checksum mismatch path=%s trace_id=%s", audioPath, telemetry.TraceID(ctx))
}

func setSongAudioResponseHeaders(w http.ResponseWriter, size int64, mime string) {
	if size > 0 {
		w.Header().Set("Content-Length", strconv.FormatInt(size, 10))
	}

	if mime != "" {
		w.Header().Set("Content-Type", mime)
		return
	}

	w.Header().Set("Content-Type", "audio/mpeg")
}

func logSecurityEvent(ctx context.Context, event string, details string) {
	logging.Securityf(ctx, "security_event=%s %s", event, details)
}
