package main

type MediaInitResponse struct {
	MediaId int64 `json:"media_id"`
}

type MediaStatusResponse struct {
	MediaId        int64  `json:"media_id"`
	MediaIdString  string `json:"media_id_string"`
	ProcessingInfo struct {
		CheckAfterSecs  int64  `json:"check_after_secs,omitempty"`
		ProgressPercent int64  `json:"progress_percent,omitempty"`
		State           string `json:"state"`
		Error           struct {
			Code    int    `json:"code"`
			Name    string `json:"name"`
			Message string `json:"message"`
		} `json:"error,omitempty"`
	} `json:"processing_info"`
}

type MediaFinalizeResponse struct {
	MediaId          int64  `json:"media_id"`
	MediaIdString    string `json:"media_id_string"`
	ExpiresAfterSecs int    `json:"expires_after_secs"`
	Size             int    `json:"size"`
	ProcessingInfo   *struct {
		State          string `json:"state"`
		CheckAfterSecs int    `json:"check_after_secs"`
	} `json:"processing_info"`
}
