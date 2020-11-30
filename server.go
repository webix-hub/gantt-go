package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/jinzhu/configor"
	"github.com/unrolled/render"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var format = render.New()

type Response struct {
	Invalid bool   `json:"invalid"`
	Error   string `json:"error"`
	ID      string `json:"id"`
}

type TaskInfo struct {
	ID        int     `json:"id"`
	Text      string  `json:"text"`
	StartDate string  `db:"start_date" json:"start_date"`
	Type      string  `json:"type"`
	Duration  int     `json:"duration"`
	Parent    int     `json:"parent"`
	Progress  float32 `json:"progress"`
	Open      int     `db:"opened" json:"open"`
	Details   string  `json:"details"`
}

type LinkInfo struct {
	ID     int `json:"id"`
	Source int `json:"source"`
	Target int `json:"target"`
	Type   int `json:"type"`
}

var conn *sqlx.DB

type AppConfig struct {
	Port         string
	ResetOnStart bool

	DB DBConfig
}

type DBConfig struct {
	Host     string `default:"localhost"`
	Port     string `default:"3306"`
	User     string `default:"root"`
	Password string `default:"1"`
	Database string `default:"projects"`
}

var Config AppConfig

func main() {
	flag.StringVar(&Config.Port, "port", ":3000", "port for web server")
	flag.Parse()

	configor.New(&configor.Config{ENVPrefix: "APP", Silent: true}).Load(&Config, "config.yml")

	// common drive access
	var err error

	connStr := fmt.Sprintf("%s:%s@(%s:%s)/%s?multiStatements=true&parseTime=true",
		Config.DB.User, Config.DB.Password, Config.DB.Host, Config.DB.Port, Config.DB.Database)
	conn, err = sqlx.Connect("mysql", connStr)
	if err != nil {
		log.Fatal(err)
	}

	migration(conn)

	if err != nil {
		log.Fatal(err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	cors := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	})
	r.Use(cors.Handler)

	r.Get("/tasks", func(w http.ResponseWriter, r *http.Request) {
		data := make([]TaskInfo, 0)

		err := conn.Select(&data, "SELECT task.* FROM task ORDER BY start_date;")
		if err != nil {
			format.Text(w, 500, err.Error())
			return
		}

		format.JSON(w, 200, data)
	})

	r.Put("/tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		r.ParseForm()

		err = sendUpdateQuery("task", r.Form, id)
		if err != nil {
			format.Text(w, 500, err.Error())
			return
		}

		format.JSON(w, 200, Response{ID: id})
	})

	r.Delete("/tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		_, err := conn.Exec("DELETE FROM task WHERE id = ? OR parent = ?", id, id)
		if err != nil {
			format.Text(w, 500, err.Error())
			return
		}
		_, err = conn.Exec("DELETE FROM link WHERE source = ? OR target = ?", id, id)
		if err != nil {
			format.Text(w, 500, err.Error())
			return
		}

		format.JSON(w, 200, Response{ID: id})
	})

	r.Post("/tasks", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()

		res, err := sendInsertQuery("task", r.Form)
		if err != nil {
			format.Text(w, 500, err.Error())
			return
		}

		id, _ := res.LastInsertId()
		format.JSON(w, 200, Response{ID: strconv.FormatInt(id, 10)})
	})

	r.Get("/links", func(w http.ResponseWriter, r *http.Request) {
		data := make([]LinkInfo, 0)
		err := conn.Select(&data, "SELECT link.* FROM link")

		if err != nil {
			format.Text(w, 500, err.Error())
			return
		}

		format.JSON(w, 200, data)
	})

	r.Put("/links/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		r.ParseForm()

		err := sendUpdateQuery("link", r.Form, id)
		if err != nil {
			format.Text(w, 500, err.Error())
			return
		}

		format.JSON(w, 200, Response{ID: id})
	})

	r.Delete("/links/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		_, err := conn.Exec("DELETE FROM link WHERE id = ?", id)
		if err != nil {
			format.Text(w, 500, err.Error())
			return
		}

		format.JSON(w, 200, Response{ID: id})
	})

	r.Post("/links", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()

		res, err := sendInsertQuery("link", r.Form)
		if err != nil {
			format.Text(w, 500, err.Error())
			return
		}

		id, _ := res.LastInsertId()
		format.JSON(w, 200, Response{ID: strconv.FormatInt(id, 10)})
	})

	log.Printf("Starting webserver at port " + Config.Port)
	http.ListenAndServe(Config.Port, r)
}

// both task and link tables
var whitelistTask = []string{
	"text",
	"start_date",
	"duration",
	"parent",
	"progress",
	"open", /* ! */
	"details",
	"type",
}
var whitelistLink = []string{
	"source",
	"target",
	"type",
}

func getWhiteList(table string) []string {
	if table == "task" {
		return whitelistTask
	}
	return whitelistLink
}

func sendUpdateQuery(table string, form url.Values, id string) error {
	qs := "UPDATE " + table + " SET "
	params := make([]interface{}, 0)

	allowedFields := getWhiteList(table)
	for _, key := range allowedFields {
		value, ok := form[key]
		if ok {
			// OPEN is a reserved word in MySQL
			if key == "open" {
				key = "opened"
			}

			qs += key + " = ?, "
			params = append(params, value[0])
		}
	}
	params = append(params, id)

	_, err := conn.Exec(qs[:len(qs)-2]+" WHERE id = ?", params...)
	return err
}

func sendInsertQuery(table string, form map[string][]string) (sql.Result, error) {
	qsk := "INSERT INTO " + table + " ("
	qsv := "VALUES ("
	params := make([]interface{}, 0)

	allowedFields := getWhiteList(table)
	for _, key := range allowedFields {
		value, ok := form[key]
		if ok {
			// OPEN is a reserved word in MySQL
			if key == "open" {
				key = "opened"
			}

			qsk += key + ", "
			qsv += "?, "
			params = append(params, value[0])
		}
	}

	qsk = qsk[:len(qsk)-2] + ") "
	qsv = qsv[:len(qsv)-2] + ")"

	res, err := conn.Exec(qsk+qsv, params...)
	return res, err
}
