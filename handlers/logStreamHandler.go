package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sash2721/Relay/services"
)

type LogStreamHandler struct {
	LogStreamer *services.LogStreamer
}

func (l *LogStreamHandler) HandlerLogStream(w http.ResponseWriter, r *http.Request) {
	// extracting the deploymentID from the URL
	deploymentID := chi.URLParam(r, "deploymentID")

	// check if the build is already completed
	completeStatus := l.LogStreamer.IsCompleted(deploymentID)

	if completeStatus {
		// returning the full logs as JSON array
		completeLogs := l.LogStreamer.GetLogs(deploymentID)
		logsJsonData, err := json.Marshal(completeLogs)

		if err != nil {
			http.Error(w, `{"message":"Failed to serialize logs"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(logsJsonData)
		return
	}

	// build is in progress — open SSE connection
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, `{"message":"Streaming not supported"}`, http.StatusInternalServerError)
		return
	}

	// subscribe to the logstream for getting the logs
	ch := l.LogStreamer.Subscribe(deploymentID)
	defer l.LogStreamer.Unsubscribe(deploymentID, ch)

	for msg := range ch {
		fmt.Fprintf(w, "%s\n", msg)
		flusher.Flush()
	}
}
