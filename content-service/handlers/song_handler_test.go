package handlers

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func withDurationDetector(t *testing.T, fn func(multipart.File, string) (*int, error)) {
	t.Helper()
	prev := audioDurationDetector
	audioDurationDetector = fn
	t.Cleanup(func() {
		audioDurationDetector = prev
	})
}

type multipartTestFile struct {
	fieldName string
	fileName  string
	content   []byte
}

func buildMultipartRequest(t *testing.T, files []multipartTestFile) *http.Request {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	for _, f := range files {
		part, err := writer.CreateFormFile(f.fieldName, f.fileName)
		if err != nil {
			t.Fatalf("create form file: %v", err)
		}
		if _, err := part.Write(f.content); err != nil {
			t.Fatalf("write form file: %v", err)
		}
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/songs", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func minimalWavBytes(payload []byte) []byte {
	l := len(payload)
	if l > math.MaxUint32-44 {
		panic("payload too large for minimal WAV header")
	}

	dataSize := uint32(l)
	chunkSize := uint32(36) + dataSize
	byteRate := uint32(8000 * 2) // 8kHz * mono * 16-bit
	blockAlign := uint16(2)

	buf := make([]byte, 44+len(payload))
	copy(buf[0:4], "RIFF")
	binary.LittleEndian.PutUint32(buf[4:8], chunkSize)
	copy(buf[8:12], "WAVE")
	copy(buf[12:16], "fmt ")
	binary.LittleEndian.PutUint32(buf[16:20], 16)
	binary.LittleEndian.PutUint16(buf[20:22], 1) // PCM
	binary.LittleEndian.PutUint16(buf[22:24], 1) // mono
	binary.LittleEndian.PutUint32(buf[24:28], 8000)
	binary.LittleEndian.PutUint32(buf[28:32], byteRate)
	binary.LittleEndian.PutUint16(buf[32:34], blockAlign)
	binary.LittleEndian.PutUint16(buf[34:36], 16)
	copy(buf[36:40], "data")
	binary.LittleEndian.PutUint32(buf[40:44], dataSize)
	copy(buf[44:], payload)

	return buf
}

func TestParseMultipartWithLimit_InvalidMultipart(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/songs", strings.NewReader("not-a-multipart-body"))
	req.Header.Set("Content-Type", "multipart/form-data; boundary=abc")
	rec := httptest.NewRecorder()

	err := parseMultipartWithLimit(rec, req)
	if err == nil {
		t.Fatal("expected parseMultipartWithLimit to return an error")
	}
	if !errors.Is(err, errInvalidMultipartForm) {
		t.Fatalf("expected errInvalidMultipartForm, got: %v", err)
	}
}

func TestGetSingleValidatedAudioUpload_ValidAudio_ReturnsFullReader(t *testing.T) {
	withDurationDetector(t, func(_ multipart.File, _ string) (*int, error) {
		v := 2
		return &v, nil
	})

	original := minimalWavBytes([]byte("track-bytes-after-header"))
	req := buildMultipartRequest(t, []multipartTestFile{
		{fieldName: "file", fileName: "track.wav", content: original},
	})
	rec := httptest.NewRecorder()

	if err := parseMultipartWithLimit(rec, req); err != nil {
		t.Fatalf("parseMultipartWithLimit failed: %v", err)
	}

	file, reader, ext, mime, lengthSeconds, ok := getSingleValidatedAudioUpload(rec, req)
	if !ok {
		t.Fatalf("expected valid upload, got status=%d body=%s", rec.Code, rec.Body.String())
	}
	defer file.Close()

	if ext != ".wav" {
		t.Fatalf("expected .wav extension, got %q", ext)
	}
	if mime != "audio/wav" && mime != "audio/x-wav" {
		t.Fatalf("expected wav mime, got %q", mime)
	}
	if lengthSeconds == nil || *lengthSeconds < 1 {
		t.Fatalf("expected derived duration to be set, got %+v", lengthSeconds)
	}

	got, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed reading returned reader: %v", err)
	}
	if !bytes.Equal(got, original) {
		t.Fatalf("expected reconstructed reader to contain original bytes")
	}
}

func TestGetSingleValidatedAudioUpload_RejectsMultipleFiles(t *testing.T) {
	req := buildMultipartRequest(t, []multipartTestFile{
		{fieldName: "file", fileName: "a.wav", content: minimalWavBytes([]byte("a"))},
		{fieldName: "file", fileName: "b.wav", content: minimalWavBytes([]byte("b"))},
	})
	rec := httptest.NewRecorder()

	if err := parseMultipartWithLimit(rec, req); err != nil {
		t.Fatalf("parseMultipartWithLimit failed: %v", err)
	}

	_, _, _, _, _, ok := getSingleValidatedAudioUpload(rec, req)
	if ok {
		t.Fatal("expected validation to fail for multiple files")
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "exactly one file") {
		t.Fatalf("expected exactly-one-file message, got body=%s", rec.Body.String())
	}
}

func TestGetSingleValidatedAudioUpload_RejectsMissingFileField(t *testing.T) {
	req := buildMultipartRequest(t, []multipartTestFile{
		{fieldName: "audio", fileName: "track.wav", content: minimalWavBytes([]byte("x"))},
	})
	rec := httptest.NewRecorder()

	if err := parseMultipartWithLimit(rec, req); err != nil {
		t.Fatalf("parseMultipartWithLimit failed: %v", err)
	}

	_, _, _, _, _, ok := getSingleValidatedAudioUpload(rec, req)
	if ok {
		t.Fatal("expected validation to fail for missing 'file' field")
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "field 'file'") {
		t.Fatalf("expected missing file field message, got body=%s", rec.Body.String())
	}
}

func TestGetSingleValidatedAudioUpload_RejectsNonAudioMime(t *testing.T) {
	pngHeader := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}
	req := buildMultipartRequest(t, []multipartTestFile{
		{fieldName: "file", fileName: "image.png", content: append(pngHeader, []byte("not-audio")...)},
	})
	rec := httptest.NewRecorder()

	if err := parseMultipartWithLimit(rec, req); err != nil {
		t.Fatalf("parseMultipartWithLimit failed: %v", err)
	}

	_, _, _, _, _, ok := getSingleValidatedAudioUpload(rec, req)
	if ok {
		t.Fatal("expected validation to fail for non-audio file")
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "file is not a valid audio file") {
		t.Fatalf("expected non-audio message, got body=%s", rec.Body.String())
	}
}
