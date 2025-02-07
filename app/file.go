package app

import (
	"github.com/slack-go/slack"
	"github.com/sourcegraph/conc/pool"
)

// AsyncUpload async UploadFileV2
// Required scopes : `files:write`
func (c *Client) AsyncUpload(params slack.UploadFileV2Parameters) *pool.ResultErrorPool[*slack.FileSummary] {
	p := pool.NewWithResults[*slack.FileSummary]().WithErrors()

	p.Go(func() (file *slack.FileSummary, err error) {
		file, err = c.api.UploadFileV2(params)
		if err != nil {
			return nil, err
		}
		return file, nil
	})

	return p
}
