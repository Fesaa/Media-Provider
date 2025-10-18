// This file contains all methods required from publication to communicate with the frontend
// Helper methods may be used by publication elsewhere, but should generally be for communication
// first and foremost

package publication

import (
	"encoding/json"
	"fmt"
	"math"
	"slices"
	"strings"

	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/Fesaa/Media-Provider/utils"
)

func (p *publication) GetInfo() payload.InfoStat {
	refUrl := func() string {
		if p.series != nil {
			return p.series.RefUrl
		}
		return ""
	}()

	return payload.InfoStat{
		Provider:     p.req.Provider,
		Id:           p.Id(),
		ContentState: p.state,
		Name:         p.Title(),
		RefUrl:       refUrl,
		Size:         p.readableSize(),
		Downloading:  p.state == payload.ContentStateDownloading,
		Progress:     int64(math.Floor(p.speedTracker.Progress())),
		SpeedType:    payload.IMAGES,
		Speed:        int64(math.Floor(p.speedTracker.IntermediateSpeed())),
		//Estimated:    int64(p.speedTracker.EstimatedTimeRemaining()),
		DownloadDir: p.GetDownloadDir(),
	}
}

func (p *publication) readableSize() string {
	if len(p.toDownloadUserSelected) == 0 {
		return fmt.Sprintf("%d Chapters", len(p.toDownload))
	}

	return fmt.Sprintf("%d Chapters", len(p.toDownloadUserSelected))
}

func (p *publication) Message(message payload.Message) (payload.Message, error) {
	var resp json.RawMessage
	var err error

	switch message.MessageType {
	case payload.MessageListContent:
		resp, err = json.Marshal(p.ContentList())
	case payload.SetToDownload:
		err = p.SetUserFiltered(message.Data)
	case payload.StartDownload:
		err = p.MarkReady()
	default:
		err = services.ErrUnknownMessageType
	}

	if err != nil {
		return payload.Message{}, err
	}

	return payload.Message{
		Provider:    p.Provider(),
		ContentId:   p.Id(),
		MessageType: message.MessageType,
		Data:        resp,
	}, nil
}

func (p *publication) MarkReady() error {
	if p.state != payload.ContentStateWaiting {
		return services.ErrWrongState
	}

	p.SetState(payload.ContentStateReady)
	return p.client.MoveToDownloadQueue(p.Id())
}

func (p *publication) SetUserFiltered(msg json.RawMessage) error {
	if p.state != payload.ContentStateWaiting &&
		p.state != payload.ContentStateReady {
		return services.ErrWrongState
	}

	var filter []string
	err := json.Unmarshal(msg, &filter)
	if err != nil {
		return err
	}
	p.toDownloadUserSelected = filter
	p.signalR.SizeUpdate(p.req.OwnerId, p.Id(), p.readableSize())
	return nil
}

func (p *publication) ContentList() []payload.ListContentData {
	if p.series == nil {
		return []payload.ListContentData{}
	}

	chapters := p.series.Chapters
	if len(chapters) == 0 {
		return []payload.ListContentData{}
	}

	data := utils.GroupBy(chapters, func(v Chapter) string {
		return v.Volume
	})

	childrenFunc := func(chapters []Chapter) []payload.ListContentData {
		slices.SortFunc(chapters, func(a, b Chapter) int {
			if a.Volume != b.Volume {
				return (int)(b.VolumeFloat() - a.VolumeFloat())
			}
			return (int)(b.ChapterFloat() - a.ChapterFloat())
		})

		return utils.Map(chapters, func(chapter Chapter) payload.ListContentData {
			return payload.ListContentData{
				SubContentId: chapter.Id,
				Selected:     p.willBeDownloaded(chapter),
				Label:        strings.TrimSpace(chapter.Label()),
			}
		})
	}

	sortSlice := utils.Keys(data)
	slices.SortFunc(sortSlice, utils.SortFloats)

	out := make([]payload.ListContentData, 0, len(data))
	for _, volume := range sortSlice {
		chaptersInVolume := data[volume]

		// Do not add No Volume label if there are no volumes
		if volume == "" && len(sortSlice) == 1 {
			out = append(out, childrenFunc(chaptersInVolume)...)
			continue
		}

		out = append(out, payload.ListContentData{
			Label:    utils.Ternary(volume == "", "No Volume", fmt.Sprintf("Volume %s", volume)),
			Children: childrenFunc(chaptersInVolume),
		})
	}
	return out
}

func (p *publication) willBeDownloaded(chapter Chapter) bool {
	if len(p.toDownloadUserSelected) > 0 {
		return slices.Contains(p.toDownloadUserSelected, chapter.Id)
	}

	return utils.Find(p.toDownload, func(id string) bool {
		return id == chapter.Id
	}) != nil
}
