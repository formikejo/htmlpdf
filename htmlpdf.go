package htmlpdf

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os/exec"
	"strconv"
	"time"
)

type PDFCreator interface {

	// Generate a PDF from a remote HTTP or HTTPS URL
	GenerateFromURL(url *url.URL, out io.Writer) error
}

// NewWkhtmltopdf returns a PDFCreator that uses the wkhtmltopdf backend
func NewWkhtmltopdf() (PDFCreator, error) {
	return &wkhtmltopdf{}, nil
}

func execute(cmd *exec.Cmd, timeout time.Duration) error {
	err := cmd.Start()
	if err != nil {
		return err
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-time.After(timeout):
		log.Printf("Timeout occurred, killing %s", cmd)
		if err := cmd.Process.Kill(); err != nil {
			log.Fatal("Failed to kill: ", err)
		}
		<-done
		return fmt.Errorf("Timeout")

	case err := <-done:
		return err
	}
}

type wkhtmltopdf struct{}

func (w *wkhtmltopdf) GenerateFromURL(url *url.URL, out io.Writer) error {
	cmd := exec.Command("wkhtmltopdf", url.String(), "--print-media-type", "-")
	cmd.Stdout = out

	err := execute(cmd, 3*time.Second)

	return err
}

func (w *wkhtmltopdf) GenerateFromHtml(html io.Reader, out io.Writer) error {
	return nil
}

type PDFHandler struct {
	Creator PDFCreator
}

func (h *PDFHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	u, err := url.Parse(r.URL.Query().Get("url"))
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	log.Printf("Rendering %s", u)

	b := new(bytes.Buffer)
	err = h.Creator.GenerateFromURL(u, b)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	log.Printf("Rendered %d bytes", b.Len())

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Length", strconv.Itoa(b.Len()))
	b.WriteTo(w)
}
