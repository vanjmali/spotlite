package handlers

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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
	header := []byte{
		'R', 'I', 'F', 'F',
		0x24, 0x00, 0x00, 0x00,
		'W', 'A', 'V', 'E',
		'f', 'm', 't', ' ',
	}
	return append(header, payload...)
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
	original := minimalWavBytes([]byte("track-bytes-after-header"))
	req := buildMultipartRequest(t, []multipartTestFile{
		{fieldName: "file", fileName: "track.wav", content: original},
	})
	rec := httptest.NewRecorder()

	if err := parseMultipartWithLimit(rec, req); err != nil {
		t.Fatalf("parseMultipartWithLimit failed: %v", err)
	}

	file, reader, ext, mime, ok := getSingleValidatedAudioUpload(rec, req)
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

	_, _, _, _, ok := getSingleValidatedAudioUpload(rec, req)
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

	_, _, _, _, ok := getSingleValidatedAudioUpload(rec, req)
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

	_, _, _, _, ok := getSingleValidatedAudioUpload(rec, req)
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
