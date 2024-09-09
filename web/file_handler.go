package web

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/phuslu/log"

	"github.com/peteraba/cloudy-files/apperr"
	"github.com/peteraba/cloudy-files/service"
)

type FileHandler struct {
	sessionService *service.Session
	fileService    *service.File
	logger         *log.Logger
}

func NewFileHandler(sessionService *service.Session, fileService *service.File, logger *log.Logger) *FileHandler {
	return &FileHandler{
		sessionService: sessionService,
		fileService:    fileService,
		logger:         logger,
	}
}

// SetupRoutes sets up the HTTP server.
func (fh *FileHandler) SetupRoutes(mux *http.ServeMux) *http.ServeMux {
	mux.HandleFunc("GET /files", fh.ListFiles)
	mux.HandleFunc("DELETE /files/{id}", fh.DeleteFile)
	mux.HandleFunc("POST /file-uploads", fh.UploadFile)
	mux.HandleFunc("GET /file-uploads", fh.RetrieveFile)

	return mux
}

// ListFiles lists files.
func (fh *FileHandler) ListFiles(w http.ResponseWriter, r *http.Request) {
	files, err := fh.fileService.List(r.Context(), nil, true)
	if err != nil {
		problem(w, r, err, fh.logger)

		return
	}

	if IsJSONRequest(r) {
		sendJSON(w, files, fh.logger)

		return
	}

	fileHTML := make([]string, 0, len(files))
	for _, file := range files {
		fileHTML = append(fileHTML, fmt.Sprintf(
			`<tr>
	<td>%s</td>
	<td>%s</td>
</tr>
`,
			file.Name,
			strings.Join(file.Access, ", "),
		))
	}

	tmpl := fmt.Sprintf(
		`<table>
	<thead>
		<tr>
			<th>Name</th>
			<th>Access</th>
		</tr>
	</thead>
	<tbody>
%s
	</tbody>
</table>
`,
		strings.Join(fileHTML, ""),
	)

	sendHTML(w, tmpl, fh.logger)
}

// DeleteFile deletes a file.
func (fh *FileHandler) DeleteFile(w http.ResponseWriter, r *http.Request) {
	problem(w, r, apperr.ErrNotImplemented, fh.logger)
}

// UploadFile uploads a file.
func (fh *FileHandler) UploadFile(w http.ResponseWriter, r *http.Request) {
	problem(w, r, apperr.ErrNotImplemented, fh.logger)
}

// RetrieveFile retrieves a file.
func (fh *FileHandler) RetrieveFile(w http.ResponseWriter, r *http.Request) {
	problem(w, r, apperr.ErrNotImplemented, fh.logger)
}
