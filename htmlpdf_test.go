package htmlpdf_test

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/formikejo/htmlpdf"
)

var (
	wkhtmltopdf htmlpdf.PDFCreator
	wError      error
)

func init() {
	wkhtmltopdf, wError = htmlpdf.NewWkhtmltopdf()
}

func checkWkhtmltopdf(t *testing.T) {
	if wError != nil {
		t.Skip("Could not load wkhtmltopdf", wError)
	}
}

func newTestServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "<html><body>Hello world</body></html>")
	}))
}

func TestGenerateHtml(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	checkWkhtmltopdf(t)

	b := new(bytes.Buffer)
	url, _ := url.Parse(ts.URL)
	err := wkhtmltopdf.GenerateFromURL(url, b)
	if err != nil {
		t.Fatal("Could not generate PDF", err)
	}

	if b.Len() == 0 {
		t.Errorf("No PDF generated")
	}
}

type mockCreator struct {
	expected string
	output   string
}

func (m *mockCreator) GenerateFromURL(url *url.URL, out io.Writer) error {
	if url.String() != m.expected {
		return fmt.Errorf("Wrong url, expected %s but got %s", m.expected, url)
	}
	out.Write([]byte(m.output))

	return nil
}

func TestHandler(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	handler := &htmlpdf.PDFHandler{
		Creator: &mockCreator{
			expected: ts.URL,
			output:   "correct",
		},
	}

	req, err := http.NewRequest("GET", "http://request/?url="+ts.URL, nil)
	if err != nil {
		log.Fatal(err)
	}

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Wrong status code, expected HTTP 200 but got %d", w.Code)
	}
	if w.Body.String() != "correct" {
		t.Errorf("Wrong output, expected 'correct' but got %s", w.Body.String())
	}
	if w.HeaderMap.Get("Content-Type") != "application/pdf" {
		t.Errorf("Wrong mime type, expected application/pdf but got %s", w.HeaderMap.Get("Content-Type"))
	}
	if w.HeaderMap.Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("Wrong CORS header, expected * but got %s", w.HeaderMap.Get("Access-Control-Allow-Origin"))
	}
}
