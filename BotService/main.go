package main

import (
    "fmt"
    "os"
    "io/ioutil"
    "encoding/csv"
    "log"
    "strings"
    "strconv"
    "database/sql"
    
    _ "github.com/lib/pq"
    "github.com/NicoNex/echotron/v3"
)

var bot_token string
var bot_password string
var api_keys []string


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

func main() {
    fmt.Print(api_keys)
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

	for update := range echotron.PollingUpdates(bot_token) {
        /* эта проверка помогает боту не сломаться, если он получит
           неожиданный апдейт, который не получится обработать */
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
                    //MAIN FUNCTIONALITY
                    api.SendMessage("OKAY LEGO", update.ChatID(), nil)
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
