package logic

import (
	"strings"
	"translator-worker/internal/domain"
)

type AlignmentPair struct {
	SrcSegment string `json:"src"`
	TgtSegment string `json:"tgt"`
}

func MapTimestamps(originalWords []domain.SubtitleItem, alignment []AlignmentPair) []domain.SubtitleItem {
	var finalSubtitles []domain.SubtitleItem

	cursor := 0

	for _, pair := range alignment {

		srcClean := cleanText(pair.SrcSegment)
		srcWords := strings.Fields(srcClean)

		if len(srcWords) == 0 {
			continue
		}

		firstWord := srcWords[0]
		lastWord := srcWords[len(srcWords)-1]

		startTime := 0.0
		endTime := 0.0
		foundStart := false

		for i := cursor; i < len(originalWords); i++ {
			originalClean := cleanText(originalWords[i].Text)

			if strings.Contains(originalClean, firstWord) {
				startTime = originalWords[i].Start
				cursor = i
				foundStart = true
				break
			}
		}

		if foundStart {
			tempCursor := cursor
			for i := tempCursor; i < len(originalWords); i++ {
				originalClean := cleanText(originalWords[i].Text)
				if strings.Contains(originalClean, lastWord) {
					endTime = originalWords[i].End

					cursor = i + 1
					break
				}
			}
		}

		if startTime == 0 && len(finalSubtitles) > 0 {
			startTime = finalSubtitles[len(finalSubtitles)-1].End
		}
		if endTime == 0 || endTime <= startTime {
			endTime = startTime + 2.0
		}

		finalSubtitles = append(finalSubtitles, domain.SubtitleItem{
			Text:  pair.TgtSegment,
			Start: startTime,
			End:   endTime,
			Conf:  1.0,
		})
	}

	return finalSubtitles
}

func cleanText(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, ".", "")
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, "?", "")
	s = strings.ReplaceAll(s, "!", "")
	return strings.TrimSpace(s)
}
