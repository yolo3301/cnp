package skeleton

import (
	"context"
	"net/http"
	"strings"

	"github.com/yolo3301/cnp/pkg/streamer"
)

type Agent struct{}

func (a *Agent) OnNotification(ctx context.Context, req *streamer.StreamNotificationRequest) (*streamer.StreamNotificationResult, error) {
	filename := retrieveFilename(req)

	if isDownload(req) {
		return &streamer.StreamNotificationResult{
			Response: &streamer.HttpResponseInfo{
				HttpStatus: http.StatusOK,
				Payload: &streamer.HttpResponseInfo_Payload{
					FileObject: &streamer.HttpResponseInfo_Payload_FileObject{
						Path: filename,
					},
				},
			},
		}, nil
	}

	if req.GetFirst() {
		return &streamer.StreamNotificationResult{
			DropTarget: &streamer.DropTarget{
				FileTarget: &streamer.DropTarget_FileTarget{
					Path: filename,
				},
			},
		}, nil
	}

	if req.GetFinalStatus() == int32(streamer.StreamNotificationRequest_BAD_REQUEST) {
		return &streamer.StreamNotificationResult{
			Response: &streamer.HttpResponseInfo{
				HttpStatus: http.StatusBadRequest,
			},
		}, nil
	}

	if req.GetFinalStatus() == int32(streamer.StreamNotificationRequest_INTERNAL_ERROR) {
		return &streamer.StreamNotificationResult{
			Response: &streamer.HttpResponseInfo{
				HttpStatus: http.StatusInternalServerError,
			},
		}, nil
	}

	return &streamer.StreamNotificationResult{
		Response: &streamer.HttpResponseInfo{
			HttpStatus: http.StatusCreated,
		},
	}, nil
}

// TODO: use URL to match.
func retrieveFilename(req *streamer.StreamNotificationRequest) string {
	parts := strings.Split(req.GetRequest().GetReqUri(), "/")
	return parts[len(parts)-1]
}

func isDownload(req *streamer.StreamNotificationRequest) bool {
	return req.GetRequest().GetMethod() == http.MethodGet
}
