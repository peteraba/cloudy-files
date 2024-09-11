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
	service *service.File
	cookie  *service.Cookie
	logger  *log.Logger
}

func NewFileHandler(fileService *service.File, cookie *service.Cookie, logger *log.Logger) *FileHandler {
	return &FileHandler{
		service: fileService,
		cookie:  cookie,
		logger:  logger,
	}
}

// ListFiles lists files.
func (fh *FileHandler) ListFiles(w http.ResponseWriter, r *http.Request) {
	userSession, err := fh.cookie.GetSessionUser(r)
	if err != nil {
		Problem(w, fh.logger, err)

		return
	}

	if !userSession.IsAdmin {
		Problem(w, fh.logger, apperr.ErrAccessDenied)

		return
	}

	files, err := fh.service.List(r.Context(), nil, true)
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

	Send(w, tmpl)
}
