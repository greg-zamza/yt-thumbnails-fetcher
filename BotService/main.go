package main

import (
    "os"
    "time"
    "io/ioutil"
    "strings"
    "strconv"
    "fmt"
    "log"
    "encoding/csv"
    "encoding/json"
    "database/sql"
    "net/http"
    "net/url"
    
    _ "github.com/lib/pq"
    "github.com/NicoNex/echotron/v3"
)

var bot_token string
var bot_password string
var api_keys []string
const apiURL string = "https://www.googleapis.com/youtube/v3/search"
const filterBaseURL string = "http://FilterService:8081/process"

func init() {
    /* декларируем секретные данные */
	content, err := ioutil.ReadFile("/run/secrets/bot_password")
	if err != nil {
		log.Fatal(err)
	}
	bot_password = strings.TrimRight(string(content), "\n")

    content, err = ioutil.ReadFile("/run/secrets/bot_token")
	if err != nil {
		log.Fatal(err)
	}
	bot_token = strings.TrimRight(string(content), "\n")

    file, err := os.Open("/run/secrets/yt_api_keys")
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()

    reader := csv.NewReader(file)
    tmp_api_keys, err := reader.ReadAll()
    if err != nil {
        log.Fatal(err)
    }

    for _, row := range tmp_api_keys {
        api_keys = append(api_keys, row[0])
    }
}

func current_token(db *sql.DB) string {
/* проверяет таблицу next_page в базе данных.
 * Если с даты в поле refresh прошло больше
 * 5 дней, возвращает пустую строку, если 
 * меньше - возвращает nextPageToken
 */
    var date time.Time
    var nextPageToken string
        
    err := db.QueryRow("SELECT token, refresh FROM next_page WHERE id=$1", 1).Scan(&nextPageToken, &date)
    if err != nil {
        log.Println("error in func current_token")
        return ""
    }

    elapsed := time.Since(date)
    days := int(elapsed.Hours() / 24)
    if days <= 5 {
        return nextPageToken
    }
    return ""
} 

func indexOf(slice []string, element string) int {
	for i, v := range slice {
		if v == element {
			return i
		}
	}
	return -1
}

// структура, описывающая response от yt data api (search)
type YouTubeResponse struct {
	NextPageToken string `json:"nextPageToken"`
	Items []struct {
		ID   struct {
			Kind    string `json:"kind"`
			VideoID string `json:"videoId"`
		} `json:"id"`
		Snippet struct {
			ChannelID   string    `json:"channelId"`
			Title       string    `json:"title"`
			Description string    `json:"description"`
			Thumbnails  struct {
				Default struct {
					URL    string `json:"url"`
					Width  int    `json:"width"`
					Height int    `json:"height"`
				} `json:"default"`
				High struct {
					URL    string `json:"url"`
					Width  int    `json:"width"`
					Height int    `json:"height"`
				} `json:"high"`
			} `json:"thumbnails"`
			ChannelTitle         string    `json:"channelTitle"`
		} `json:"snippet"`
	} `json:"items"`
}

// структура JSON ответа от FilterService
type FilterServiceAnswer struct {
    Name            string `json:"name"`
	SubscriberCount string `json:"subscriberCount"`
	URL             string `json:"url"`
	ThumbURLa       string `json:"thumbURLa"`
	ThumbURLb       string `json:"thumbURLb"`
	ThumbURLc       string `json:"thumbURLc"`
}

// метод FilterServiceAnswer, возвращающий строку с её полями
func (a FilterServiceAnswer) String() string {
	result := ""

	addField := func(fieldName, value string) {
		if value != "" {
			result += fmt.Sprintf("%s: %s\n", fieldName, value)
		}
	}

	addField("name", a.Name)
	addField("subscribers", a.SubscriberCount)
	addField("link", a.URL)

	return result
}

func main() {
    // connecting to database
    var conn_params string = fmt.Sprintf(
        "user=%s dbname=%s sslmode=disable host=DatabaseService password=%s",
        os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_DB"), os.Getenv("POSTGRES_PASSWORD"))
    db, err := sql.Open("postgres", conn_params)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    api := echotron.NewAPI(bot_token)
    var isAdmin bool
    var counter int
    var nextPageToken string
    var yt_api_key = api_keys[0]

	for update := range echotron.PollingUpdates(bot_token) {
        // помогает боту не сломаться от невалидного апдейта
        if update.Message == nil {
            log.Println("Unhandled update")
        } else {
            err = db.QueryRow("SELECT EXISTS (SELECT 1 FROM admins WHERE id = $1)", update.Message.From.ID).Scan(&isAdmin)
            if err != nil {
                log.Println("not selected from db")
            }

            if isAdmin {
                // валидация (message must contain only int fom 1 to 50)
                n, err := strconv.Atoi(update.Message.Text)
                if err != nil || n < 1 || n > 50 {
                    api.SendMessage("Please send number from 1 to 50", update.ChatID(), nil)
                } else {
                    /* api.SendMessage(fmt.Sprintf("current api_key: ***%s***", yt_api_key[3:10]), update.ChatID(), nil) */
                    // запрос к youtube data API
                    counter = 0
                    for ;counter < n; {
                        nextPageToken = current_token(db)
	                    params := url.Values{}
	                    params.Set("part", "snippet")
                        params.Set("maxResults", "50")
                        params.Set("order", "date")
                        params.Set("regionCode", "RU")
                        params.Set("relevanceLanguage", "RU")
	                    params.Set("type", "video")
                        params.Set("pageToken", nextPageToken)
	                    params.Set("key", yt_api_key)
                        dateString := fmt.Sprintf(time.Now().Add(time.Hour * -720).Format("2006-01-02")+"T00:00:00Z")
                        params.Set("publishedAfter", dateString)

	                    requestURL := fmt.Sprintf("%s?%s", apiURL, params.Encode())
                        
                        client := &http.Client{}
    
                        req, err := http.NewRequest("GET", requestURL, nil)
	                    if err != nil {
		                    log.Println("warning")
		                    break
	                    }
                        response, err := client.Do(req)
                        if err != nil {
		                    log.Println("warning")
		                    break
	                    }
	                    defer response.Body.Close()

                        if response.StatusCode != 200 {
                            messagetext := fmt.Sprintf("yt data API returned error: %v\nПопробуй отправить запрос ещё раз", response.Status)
                            api.SendMessage(messagetext, update.ChatID(), nil)
                            if len(api_keys) == 1 {
                                api.SendMessage("It seems that the daily requests limit has been exhausted", update.ChatID(), nil)
                            } else {
                                position_api := indexOf(api_keys, yt_api_key)
                                if position_api == len(api_keys)-1 {
                                    yt_api_key = api_keys[0]
                                } else {
                                    yt_api_key = api_keys[position_api+1]
                                }
                            }
                            break
	                    }

	                    // Чтение ответа
	                    body, err := ioutil.ReadAll(response.Body)
	                    if err != nil {
		                    api.SendMessage("Ошибка при чтении ответа", update.ChatID(), nil)
		                    break
	                    }
	                    // Распаковка JSON
	                    var result YouTubeResponse
	                    err = json.Unmarshal(body, &result)
	                    if err != nil {
		                    api.SendMessage("Ошибка при распаковке JSON:", update.ChatID(), nil)
		                    break
	                    }
                        
                        // Записываем новый токен в бд
                        if nextPageToken == "" {
                            _, err := db.Exec("UPDATE next_page SET token=$1, refresh=$2 WHERE id=$3", result.NextPageToken, time.Now(), 1)
                            if err != nil {
                                log.Println(err)
                            }
                        } else {
                            _, err := db.Exec("UPDATE next_page SET token=$1 WHERE id=$2", result.NextPageToken, 1)
                            if err != nil {
                                log.Println(err)
                            }
                        }
                        
                        // Проходимся по элементам коллекции Items
                        for _, video := range result.Items {
                            var ignore bool
                            err = db.QueryRow("SELECT EXISTS (SELECT 1 FROM ignore_list WHERE id = $1)", video.Snippet.ChannelID).Scan(&ignore)
                            if err != nil {
                                log.Println("not selected from db")
                            }
                            if !ignore {
                                _, err = db.Exec("INSERT INTO ignore_list (id) VALUES ($1)", video.Snippet.ChannelID)
                                if err != nil {
                                    log.Println("not inserted to db")
                                }

                                // request to FilterService microservice
                                paramsFilter := url.Values{}
                                paramsFilter.Set("channel_id", video.Snippet.ChannelID)
                                paramsFilter.Set("key", yt_api_key)

                                filterURL := fmt.Sprintf("%s?%s", filterBaseURL, paramsFilter.Encode())
                                requestFilter, err := http.NewRequest("GET", filterURL, nil)
                                if err != nil {
                                    log.Println("request filter error (http.NewRequest)")
                                    continue
                                }

                                responseFilter, err := client.Do(requestFilter)
                                if err != nil {
                                    log.Println("request filter error (client.Do)")
                                    continue
                                }
                                defer responseFilter.Body.Close()

                                bodyFilter, err := ioutil.ReadAll(responseFilter.Body)
                                
                                if err == nil {
                                    if string(bodyFilter) != "" {
                                        var resultAnswer FilterServiceAnswer
                                        err = json.Unmarshal(bodyFilter, &resultAnswer)
                                        if err != nil {
                                            continue
                                        }

                                        var photos []echotron.InputFile
                                        if url := resultAnswer.ThumbURLa; url != "" {
	                                    	photos = append(photos, echotron.NewInputFileURL(url))
	                                    }
	                                    if url := resultAnswer.ThumbURLb; url != "" {
		                                    photos = append(photos, echotron.NewInputFileURL(url))
	                                    }
	                                    if url := resultAnswer.ThumbURLc; url != "" {
		                                    photos = append(photos, echotron.NewInputFileURL(url))
	                                    }

                                        var photosMedia []echotron.GroupableInputMedia
                                        for idx, photo := range photos {
                                            if idx == 0 {
                                                photoMedia := echotron.InputMediaPhoto{
                                                    Type: echotron.MediaTypePhoto,
                                                    Media: photo,
                                                    Caption: resultAnswer.String(),
                                                }
                                                photosMedia = append(photosMedia, echotron.GroupableInputMedia(photoMedia))
                                            } else {
                                                photoMedia := echotron.InputMediaPhoto{
                                                    Type: echotron.MediaTypePhoto,
                                                    Media: photo,
                                                }
                                                photosMedia = append(photosMedia, echotron.GroupableInputMedia(photoMedia))
                                            }
                                        }

                                        api.SendMediaGroup(update.ChatID(), photosMedia, nil)
                                        counter++
                                    }
                                }
                            }
                        }
                        /*log_message := fmt.Sprintf("log:\ntotal: %d", counter)
                        api.SendMessage(log_message, update.ChatID(), nil)*/
                    }
                }
            } else {
                if update.Message.Text == bot_password {
                    api.SendMessage("Welcome! 👋", update.ChatID(), nil)
                    _, err = db.Exec("INSERT INTO admins (id) VALUES ($1)", update.Message.From.ID)
                    if err != nil {
                        log.Println("not inserted to db")
                    }
                } else {
                    api.SendMessage("please enter the password", update.ChatID(), nil)
                }
            }
        }
	}
}
