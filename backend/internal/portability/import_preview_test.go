package portability

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestImportPreview_ReturnsDetectedColumnsAndSampleRows(t *testing.T) {
	g := gin.New()
	handler := NewHandler(NewImportPreviewService(), nil, nil, nil)
	g.POST("/import/csv", withUser("user-1"), handler.ImportPreview)

	body, contentType := buildMultipartCSV(t, "preview.csv", "date,amount,description\n2026-03-01,12.5,lunch\n2026-03-02,8.8,coffee\n")
	req := httptest.NewRequest(http.MethodPost, "/import/csv", body)
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()

	g.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rec.Code, rec.Body.String())
	}
	assertBodyContains(t, rec.Body.String(), `"columns":["date","amount","description"]`)
	assertBodyContains(t, rec.Body.String(), `"sample_rows":[["2026-03-01","12.5","lunch"],["2026-03-02","8.8","coffee"]]`)
}

func TestImportPreview_IncludesMappingSlotsForConfirmStage(t *testing.T) {
	g := gin.New()
	handler := NewHandler(NewImportPreviewService(), nil, nil, nil)
	g.POST("/import/csv", withUser("user-1"), handler.ImportPreview)

	body, contentType := buildMultipartCSV(t, "preview.csv", "date,amount,description\n2026-03-01,12.5,lunch\n")
	req := httptest.NewRequest(http.MethodPost, "/import/csv", body)
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()

	g.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rec.Code, rec.Body.String())
	}
	assertBodyContains(t, rec.Body.String(), `"mappingSlots":["amount","date","description","category","account","tag"]`)
	assertBodyContains(t, rec.Body.String(), `"mappingCandidates":{`)
}

func TestImportPreview_InvalidFile_ReturnsIMPORT_INVALID_FILE(t *testing.T) {
	g := gin.New()
	handler := NewHandler(NewImportPreviewService(), nil, nil, nil)
	g.POST("/import/csv", withUser("user-1"), handler.ImportPreview)

	body, contentType := buildMultipartCSV(t, "invalid.csv", "not,a,valid,csv\n\"broken\n")
	req := httptest.NewRequest(http.MethodPost, "/import/csv", body)
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()

	g.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusBadRequest, rec.Code, rec.Body.String())
	}
	assertBodyContains(t, rec.Body.String(), `"error_code":"IMPORT_INVALID_FILE"`)
}

func TestImportPreview_AcceptsAccessAndPAT(t *testing.T) {
	g := gin.New()
	handler := NewHandler(NewImportPreviewService(), nil, nil, nil)
	g.POST("/import/csv", withUser("user-1"), handler.ImportPreview)

	for _, tokenType := range []string{"access", "pat"} {
		body, contentType := buildMultipartCSV(t, tokenType+".csv", "date,amount\n2026-03-01,12.5\n")
		req := httptest.NewRequest(http.MethodPost, "/import/csv", body)
		req.Header.Set("Content-Type", contentType)
		rec := httptest.NewRecorder()

		g.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d for %s token, got %d body=%s", http.StatusOK, tokenType, rec.Code, rec.Body.String())
		}
	}
}

func withUser(userID string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	}
}

func buildMultipartCSV(t *testing.T, filename string, content string) (*bytes.Buffer, string) {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write([]byte(content)); err != nil {
		t.Fatalf("write form file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}
	return body, writer.FormDataContentType()
}

func assertBodyContains(t *testing.T, body string, fragment string) {
	t.Helper()
	if !bytes.Contains([]byte(body), []byte(fragment)) {
		t.Fatalf("expected body %s to contain %s", body, fragment)
	}
}
