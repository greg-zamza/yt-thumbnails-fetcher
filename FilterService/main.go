/* Сервис принимает channel_id и api_key для youtube data API
 * Возвращает json со ссылкой на канал и ссылками на превью последних трёх видео в случае, если канал валидный
 * В ином случае или в случае ошибки возвращает пустую строку
 */
package main

import (
	"fmt"
	"net/http"
    "net/url"
    "encoding/json"
    "io/ioutil"

    "strconv"
)

// структура JSON ответа от youtube data api (channels.list)
type Channel struct {
	Items []struct {
		Snippet struct {
			Title       string    `json:"title"`
			CustomURL   string    `json:"customUrl"`
			Country     string    `json:"country"`
		} `json:"snippet"`
		ContentDetails struct {
			RelatedPlaylists struct {
				Uploads string `json:"uploads"`
			} `json:"relatedPlaylists"`
		} `json:"contentDetails"`
		Statistics struct {
			SubscriberCount    string `json:"subscriberCount"`
			VideoCount         string `json:"videoCount"`
		} `json:"statistics"`
		BrandingSettings struct {
			Channel struct {
				Title          string `json:"title"`
				Description    string `json:"description"`
			} `json:"channel"`
			Image struct {
				BannerExternalURL string `json:"bannerExternalUrl"`
			} `json:"image"`
		} `json:"brandingSettings"`
	} `json:"items"`
}

// структура JSON ответа от youtube data API (Playlistitems)
type PlaylistItems struct {
	Items []struct {
		Snippet struct {
			Thumbnails struct {
				Maxres struct {
					URL    string `json:"url"`
					Width  int    `json:"width"`
					Height int    `json:"height"`
				} `json:"maxres"`
			} `json:"thumbnails"`
		} `json:"snippet"`
	} `json:"items"`
}


// структура JSON ответа, отправляемого микросервисом
type Answer struct {
    Name               string `json:"name"`
    SubscriberCount    string `json:"subscriberCount"`
    URL                string `json:"url"`
    ThumbURLa          string `json:"thumbURLa"`
    ThumbURLb          string `json:"thumbURLb"`
    ThumbURLc          string `json:"thumbURLc"`
}

const apiURL string = "https://www.googleapis.com/youtube/v3/channels"
const thumbURL string = "https://www.googleapis.com/youtube/v3/playlistItems"
func handler(w http.ResponseWriter, r *http.Request) {
    output := ""

    channelId := r.URL.Query().Get("channel_id")
    apiKey := r.URL.Query().Get("key")
    
    params := url.Values{}
    params.Set("part", "id,snippet,statistics,contentDetails")
    params.Set("id", channelId)
    params.Set("key", apiKey)

    requestURL := fmt.Sprintf("%s?%s", apiURL, params.Encode())
    
    // get response (channels)
    client := &http.Client{}
    
    req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		fmt.Fprintf(w, output)
		return
	}
    response, err := client.Do(req)
    if err != nil {
		fmt.Fprintf(w, output)
		return
	}
    
    if response.StatusCode != 200 {
        fmt.Fprintf(w, output)
        return
    }
	defer response.Body.Close()

    // read response
    body, err := ioutil.ReadAll(response.Body)
    if err != nil {
        fmt.Fprintf(w, output)
        return
    }

    // unmarshalling JSON
    var result Channel
    err = json.Unmarshal(body, &result)
    if err != nil {
        fmt.Fprintf(w, output)
        return
    }
    
    // VALIDATION (step 1)
    subscribers, err := strconv.Atoi(result.Items[0].Statistics.SubscriberCount)
    if err != nil {
        fmt.Fprintf(w, output)
        return
    }

    videoCount, err := strconv.Atoi(result.Items[0].Statistics.VideoCount)
    if err != nil {
        fmt.Fprintf(w, output)
        return
    }
    
    // валидация полей Statistics
    if subscribers > 1000000 || videoCount < 3 {
        fmt.Fprintf(w, output)
        return
    }

    // валидация поля Snippet.Country
    country := result.Items[0].Snippet.Country
    if country != "RU" && country != "UA" && country != "KZ" && country != "BY" {
        fmt.Fprintf(w, output)
        return
    }
    
    /* тут место для прочих проверок, которые когда-нибудь обязательно будут */

    // если до этого момента не произошёл return, значит канал валидный
    // и можно формировать и возвращать JSON
    
    paramsThumb := url.Values{}
    paramsThumb.Set("part", "snippet")
    paramsThumb.Set("playlistId", result.Items[0].ContentDetails.RelatedPlaylists.Uploads)
    paramsThumb.Set("key", apiKey)

    thumbRequestURL := fmt.Sprintf("%s?%s", thumbURL, paramsThumb.Encode())
    
    // get response (playlistItems)
    requestThumb, err := http.NewRequest("GET", thumbRequestURL, nil)
	if err != nil {
		fmt.Fprintf(w, output)
		return
	}
    responseThumb, err := client.Do(requestThumb)
    if err != nil {
		fmt.Fprintf(w, output)
		return
	}
    
    if responseThumb.StatusCode != 200 {
        fmt.Fprintf(w, output)
        return
    }
	defer responseThumb.Body.Close()
    
    // read response
    bodyThumb, err := ioutil.ReadAll(responseThumb.Body)
    if err != nil {
        fmt.Fprintf(w, output)
        return
    }

    // unmarshalling JSON
    var resultThumb PlaylistItems
    err = json.Unmarshal(bodyThumb, &resultThumb)
    if err != nil {
        fmt.Fprintf(w, output)
        return
    }

    // формируем JSON и отправляем его
    answer := Answer{
        Name:              result.Items[0].Snippet.Title,
        SubscriberCount:   result.Items[0].Statistics.SubscriberCount,
        URL:               "https://youtube.com/" + result.Items[0].Snippet.CustomURL,
        ThumbURLa:         resultThumb.Items[0].Snippet.Thumbnails.Maxres.URL,
        ThumbURLb:         resultThumb.Items[1].Snippet.Thumbnails.Maxres.URL,
		ThumbURLc:         resultThumb.Items[2].Snippet.Thumbnails.Maxres.URL,
    }
    
    jsonAnswer, err := json.Marshal(answer)
	if err != nil {
		fmt.Fprintf(w, output)
		return
	}
    fmt.Fprintf(w, string(jsonAnswer))
}

func main() {
	http.HandleFunc("/process", handler)
	http.ListenAndServe(":8081", nil)
}
