package handler

import (
	"zai-proxy/internal/model"
	"zai-proxy/internal/separatorrule"
)

func sanitizeUpstreamData(upstreamData *model.UpstreamData, enabled bool) {
	if !enabled || upstreamData == nil {
		return
	}

	upstreamData.Data.DeltaContent = separatorrule.Strip(upstreamData.Data.DeltaContent)
	upstreamData.Data.EditContent = separatorrule.Strip(upstreamData.Data.EditContent)
	upstreamData.Data.Content = separatorrule.Strip(upstreamData.Data.Content)
	if upstreamData.Data.Error != nil {
		upstreamData.Data.Error.Detail = separatorrule.Strip(upstreamData.Data.Error.Detail)
	}
}
