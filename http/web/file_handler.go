package web

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/phuslu/log"

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

// ListFiles lists files.
func (fh *FileHandler) ListFiles(w http.ResponseWriter, r *http.Request) {
	files, err := fh.fileService.List(r.Context(), nil, true)
	if err != nil {
		Problem(w, fh.logger, err)

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

	send(w, tmpl, fh.logger)
}
