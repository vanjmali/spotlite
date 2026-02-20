//go:build e2e

package e2e

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

const max429Retries = 8

type listResponse[T any] struct {
	Items []T `json:"items"`
}

type genreItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type artistItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type albumItem struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type songResponse struct {
	ID            string `json:"id"`
	Title         string `json:"title"`
	AudioPath     string `json:"audioPath"`
	AudioSize     int64  `json:"audioSize"`
	AudioMimeType string `json:"audioMimeType"`
	AudioChecksum string `json:"audioChecksum"`
}

func TestAudioFlowE2E(t *testing.T) {
	baseURL := os.Getenv("E2E_BASE_URL")
	if baseURL == "" {
		baseURL = "https://localhost:4443/api/content"
	}
	baseURL = strings.TrimRight(baseURL, "/")

	client := &http.Client{Timeout: 30 * time.Second}
	if err := waitHealthy(client, baseURL+"/healthz", 30*time.Second); err != nil {
		t.Fatalf("content service is not reachable at %s: %v", baseURL, err)
	}

	suffix := fmt.Sprintf("e2e-%d", time.Now().UnixNano())

	genreName := "E2E Genre " + suffix
	artistName := "E2E Artist " + suffix
	albumTitle := "E2E Album " + suffix
	songTitle := "E2E Song " + suffix

	postJSON(t, client, baseURL+"/genres", map[string]any{
		"name": genreName,
	}, http.StatusNoContent)
	genreID := findOneByField(t, client, baseURL+"/genres", "name", genreName, func(g genreItem) string { return g.ID })

	postJSON(t, client, baseURL+"/artists", map[string]any{
		"name":        artistName,
		"description": "E2E test artist",
		"genre_ids":   []string{genreID},
	}, http.StatusNoContent)
	artistID := findOneByField(t, client, baseURL+"/artists", "name", artistName, func(a artistItem) string { return a.ID })

	postJSON(t, client, baseURL+"/albums", map[string]any{
		"title":        albumTitle,
		"release_date": "2026-02-15",
		"genre_ids":    []string{genreID},
		"artist_ids":   []string{artistID},
	}, http.StatusNoContent)
	albumID := findOneByField(t, client, baseURL+"/albums", "title", albumTitle, func(a albumItem) string { return a.ID })

	audioV1 := minimalWAVBytes([]byte("first-audio-payload"))
	meta := map[string]any{
		"title":      songTitle,
		"album_id":   albumID,
		"genre_ids":  []string{genreID},
		"artist_ids": []string{artistID},
	}
	created := createSongWithAudio(t, client, baseURL+"/songs", meta, "track.wav", audioV1)
	if created.ID == "" {
		t.Fatal("expected created song id")
	}
	if created.AudioChecksum == "" {
		t.Fatal("expected audio checksum on created song")
	}
	assertSHA256Hex(t, created.AudioChecksum, audioV1)

	gotV1, status, contentType := getBytes(t, client, baseURL+"/songs/"+created.ID+"/audio")
	if status != http.StatusOK {
		t.Fatalf("expected stream status 200, got %d", status)
	}
	if !strings.Contains(contentType, "audio/") {
		t.Fatalf("expected audio content type, got %q", contentType)
	}
	if !bytes.Equal(gotV1, audioV1) {
		t.Fatal("streamed audio does not match uploaded payload (v1)")
	}

	audioV2 := minimalWAVBytes([]byte("second-audio-payload"))
	updated := putSongAudio(t, client, baseURL+"/songs/"+created.ID+"/audio", "track2.wav", audioV2)
	assertSHA256Hex(t, updated.AudioChecksum, audioV2)

	gotV2, status, _ := getBytes(t, client, baseURL+"/songs/"+created.ID+"/audio")
	if status != http.StatusOK {
		t.Fatalf("expected stream status 200 after replace, got %d", status)
	}
	if !bytes.Equal(gotV2, audioV2) {
		t.Fatal("streamed audio does not match uploaded payload (v2)")
	}
}

func waitHealthy(client *http.Client, healthURL string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		req, _ := http.NewRequest(http.MethodGet, healthURL, nil)
		res, err := client.Do(req)
		if err == nil {
			_ = res.Body.Close()
			if res.StatusCode >= 200 && res.StatusCode < 300 {
				return nil
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("health check did not pass before timeout")
}

func postJSON(t *testing.T, client *http.Client, endpoint string, payload any, expectedStatus int) {
	t.Helper()
	b, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	res, body := doRequestWith429Retry(t, client, func() (*http.Request, error) {
		req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(b))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		return req, nil
	})

	if res.StatusCode != expectedStatus {
		t.Fatalf("unexpected status %d for %s: %s", res.StatusCode, endpoint, string(body))
	}
}

func createSongWithAudio(t *testing.T, client *http.Client, endpoint string, meta map[string]any, fileName string, fileBytes []byte) songResponse {
	t.Helper()
	return uploadSongMultipart(t, client, http.MethodPost, endpoint, meta, fileName, fileBytes)
}

func putSongAudio(t *testing.T, client *http.Client, endpoint string, fileName string, fileBytes []byte) songResponse {
	t.Helper()
	return uploadSongMultipart(t, client, http.MethodPut, endpoint, nil, fileName, fileBytes)
}

func uploadSongMultipart(t *testing.T, client *http.Client, method, endpoint string, meta map[string]any, fileName string, fileBytes []byte) songResponse {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	if meta != nil {
		metaJSON, err := json.Marshal(meta)
		if err != nil {
			t.Fatalf("marshal meta: %v", err)
		}
		if err := writer.WriteField("meta", string(metaJSON)); err != nil {
			t.Fatalf("write meta field: %v", err)
		}
	}

	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write(fileBytes); err != nil {
		t.Fatalf("write file bytes: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	contentType := writer.FormDataContentType()
	rawBody := body.Bytes()
	res, respBody := doRequestWith429Retry(t, client, func() (*http.Request, error) {
		req, err := http.NewRequest(method, endpoint, bytes.NewReader(rawBody))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", contentType)
		return req, nil
	})

	if res.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status %d for %s: %s", res.StatusCode, endpoint, string(respBody))
	}

	var parsed songResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		t.Fatalf("parse response: %v; body=%s", err, string(respBody))
	}
	return parsed
}

func getBytes(t *testing.T, client *http.Client, endpoint string) ([]byte, int, string) {
	t.Helper()
	res, body := doRequestWith429Retry(t, client, func() (*http.Request, error) {
		return http.NewRequest(http.MethodGet, endpoint, nil)
	})
	return body, res.StatusCode, res.Header.Get("Content-Type")
}

func findOneByField[T any](t *testing.T, client *http.Client, endpoint, key, value string, idFn func(T) string) string {
	t.Helper()
	u := endpoint + "?" + url.Values{key: []string{value}}.Encode()

	res, body := doRequestWith429Retry(t, client, func() (*http.Request, error) {
		return http.NewRequest(http.MethodGet, u, nil)
	})
	if res.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status %d for %s: %s", res.StatusCode, u, string(body))
	}

	var parsed listResponse[T]
	if err := json.Unmarshal(body, &parsed); err != nil {
		t.Fatalf("parse list response: %v body=%s", err, string(body))
	}
	if len(parsed.Items) == 0 {
		t.Fatalf("expected at least one item from %s", u)
	}
	id := idFn(parsed.Items[0])
	if id == "" {
		t.Fatalf("empty id from %s", u)
	}
	return id
}

func minimalWAVBytes(payload []byte) []byte {
	if len(payload) > int(^uint32(0)) {
		panic("payload too large")
	}
	dataSize := uint32(len(payload))
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

func assertSHA256Hex(t *testing.T, got string, payload []byte) {
	t.Helper()
	sum := sha256.Sum256(payload)
	expected := hex.EncodeToString(sum[:])
	if got != expected {
		t.Fatalf("checksum mismatch expected=%s got=%s", expected, got)
	}
}

func doRequestWith429Retry(t *testing.T, client *http.Client, buildReq func() (*http.Request, error)) (*http.Response, []byte) {
	t.Helper()

	for attempt := 0; attempt <= max429Retries; attempt++ {
		req, err := buildReq()
		if err != nil {
			t.Fatalf("create request: %v", err)
		}
		res, err := client.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		body, _ := io.ReadAll(res.Body)
		_ = res.Body.Close()

		if res.StatusCode != http.StatusTooManyRequests {
			return res, body
		}

		if attempt == max429Retries {
			return res, body
		}
		time.Sleep(backoffFromRetryAfter(res.Header.Get("Retry-After"), attempt))
	}

	t.Fatal("unreachable")
	return nil, nil
}

func backoffFromRetryAfter(retryAfter string, attempt int) time.Duration {
	if retryAfter != "" {
		if seconds, err := strconv.Atoi(retryAfter); err == nil && seconds > 0 {
			return time.Duration(seconds) * time.Second
		}
	}

	return time.Duration(attempt+1) * 250 * time.Millisecond
}
