package wubby

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jean-jacket/grss/client"
	"github.com/jean-jacket/grss/config"
	"github.com/jean-jacket/grss/feed"
	"github.com/jean-jacket/grss/routes/registry"
)

// VodsRoute defines the Wubby recent VODs route
var VodsRoute = registry.Route{
	Path:        "/vods",
	Name:        "Recent VODs",
	Maintainers: []string{"wubby"},
	Example:     "/wubby/vods",
	Description: "Get recent Wubby VODs from parasoci.al",
	Handler:     vodsHandler,
}

type wubbyTag struct {
	Tag       string      `json:"tag"`
	Score     int         `json:"score"`
	Category  string      `json:"category"`
	Timestamp interface{} `json:"timestamp"`
}

type wubbyVod struct {
	ID        int        `json:"id"`
	Title     string     `json:"title"`
	AltTitle  string     `json:"alt_title"`
	Date      string     `json:"date"`
	Timestamp int64      `json:"timestamp"`
	URL       string     `json:"url"`
	Views     int        `json:"views"`
	Duration  int        `json:"duration"`
	Rating    float64    `json:"rating"`
	Tags      []wubbyTag `json:"tags"`
}

type nextData struct {
	Props struct {
		PageProps struct {
			HP struct {
				RecentVods []wubbyVod `json:"recent_vods"`
			} `json:"hp"`
		} `json:"pageProps"`
	} `json:"props"`
}

var nextDataRegex = regexp.MustCompile(`<script id="__NEXT_DATA__" type="application/json">([^<]+)</script>`)

func vodsHandler(c *gin.Context) (*feed.Data, error) {
	httpClient := client.New(config.C)

	html, err := httpClient.Get("https://parasoci.al/", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch page: %w", err)
	}

	matches := nextDataRegex.FindSubmatch(html)
	if len(matches) < 2 {
		return nil, fmt.Errorf("could not find __NEXT_DATA__ script tag")
	}

	var data nextData
	if err := json.Unmarshal(matches[1], &data); err != nil {
		return nil, fmt.Errorf("failed to parse __NEXT_DATA__: %w", err)
	}

	vods := data.Props.PageProps.HP.RecentVods

	sort.Slice(vods, func(i, j int) bool {
		return vods[i].Timestamp > vods[j].Timestamp
	})

	feedData := &feed.Data{
		Title:       "Wubby Recent VODs",
		Link:        "https://parasoci.al/",
		Description: "Recent VODs from Wubby on parasoci.al",
		Language:    "en",
		Item:        make([]feed.Item, 0, len(vods)),
	}

	for _, vod := range vods {
		pubDate := time.Unix(vod.Timestamp, 0)
		thumbURL := fmt.Sprintf("https://glorg.parasoci.al/thumb/%d.jpg", vod.Timestamp)

		var desc strings.Builder
		desc.WriteString(fmt.Sprintf(`<img src="%s">`, thumbURL))
		desc.WriteString("<br>")
		desc.WriteString(fmt.Sprintf("Duration: %s<br>", formatHHMMSS(int64(vod.Duration))))
		desc.WriteString("<br>")

		if len(vod.Tags) > 0 {
			desc.WriteString("Tags:<br>")

			tags := make([]wubbyTag, len(vod.Tags))
			copy(tags, vod.Tags)
			sort.Slice(tags, func(i, j int) bool {
				return tagTimestampSeconds(tags[i].Timestamp) < tagTimestampSeconds(tags[j].Timestamp)
			})

			for _, tag := range tags {
				desc.WriteString(fmt.Sprintf("%s: %s<br>", tag.Tag, formatHHMMSS(tagTimestampSeconds(tag.Timestamp))))
			}
		}

		feedData.Item = append(feedData.Item, feed.Item{
			Title:       vod.AltTitle,
			Link:        vod.URL,
			Description: desc.String(),
			PubDate:     pubDate,
			Media: &feed.Media{
				Thumbnail: &feed.MediaThumbnail{
					URL: thumbURL,
				},
			},
		})
	}

	if len(vods) > 0 {
		feedData.PubDate = time.Unix(vods[0].Timestamp, 0)
	}

	return feedData, nil
}

func formatHHMMSS(seconds int64) string {
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	secs := seconds % 60
	return fmt.Sprintf("%d:%02d:%02d", hours, minutes, secs)
}

func tagTimestampSeconds(v interface{}) int64 {
	if v == nil {
		return 0
	}
	switch t := v.(type) {
	case float64:
		return int64(t)
	case int:
		return int64(t)
	case int64:
		return t
	default:
		return 0
	}
}
