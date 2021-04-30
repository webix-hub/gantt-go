package main

import (
	"database/sql"
	"errors"
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

// Response is a general server response
type Response struct {
	ID int `json:"id"`
}

// AddResponse is a split task server response
type AddResponse struct {
	ID      int `json:"id"`
	Sibling int `json:"sibling"`
}

// TaskInfo describes a task
type TaskInfo struct {
	ID        int     `json:"id"`
	Text      string  `json:"text"`
	StartDate string  `db:"start_date" json:"start_date"`
	Type      string  `json:"type"`
	Duration  int     `json:"duration"`
	Parent    int     `json:"parent"`
	Progress  float32 `json:"progress"`
	Opened    int     `json:"opened"`
	Details   string  `json:"details"`
	Position  int     `json:"-"`
}

// LinkInfo describes a link between two tasks
type LinkInfo struct {
	ID     int `json:"id"`
	Source int `json:"source"`
	Target int `json:"target"`
	Type   int `json:"type"`
}

// Assignment describes a resource allocated to a task
type Assignment struct {
	ID       int `json:"id"`
	Task     int `json:"task"`
	Resource int `json:"resource"`
	Value    int `json:"value"`
}

// Resource describes a person or other work resource
type Resource struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	CategoryID int    `json:"category_id" db:"category_id"`
	Avatar     string `json:"avatar"`
	Unit       string `json:"unit"`
}

// Category describes a department of a company or a category of non-human resources
type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Unit string `json:"unit"`
}

var conn *sqlx.DB

// AppConfig describes settings for this backend app
type AppConfig struct {
	Port         string
	ResetOnStart bool

	DB DBConfig
}

// DBConfig describes settings for the database
type DBConfig struct {
	Host     string `default:"localhost"`
	Port     string `default:"3306"`
	User     string `default:"root"`
	Password string `default:"1"`
	Database string `default:"projects"`
}

// Config is the structure that stores the settings for this backend app
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

		err := conn.Select(&data, "SELECT task.* FROM task ORDER BY parent, position")
		if err != nil {
			format.Text(w, 500, err.Error())
			return
		}

		format.JSON(w, 200, data)
	})

	r.Put("/tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := NumberParam(r, "id")
		r.ParseForm()

		err = sendUpdateQuery("task", r.Form, id)
		if err != nil {
			format.Text(w, 500, err.Error())
			return
		}

		format.JSON(w, 200, Response{ID: id})
	})

	r.Put("/tasks/{id}/split", func(w http.ResponseWriter, r *http.Request) {
		id := NumberParam(r, "id")
		r.ParseForm()
		nid, sibling, err := splitTask(id, r.Form)
		if err != nil {
			format.Text(w, 500, err.Error())
			return
		}

		format.JSON(w, 200, AddResponse{ID: nid, Sibling: sibling})
	})

	r.Delete("/tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := NumberParam(r, "id")

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
		_, err = conn.Exec("DELETE FROM assignment WHERE task = ?", id)
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

		mode := r.Form.Get("mode")
		parent := NumberFromForm(r.Form, "parent", 0)

		err = setPosition(int(id), mode, parent)
		if err != nil {
			format.Text(w, 500, err.Error())
			return
		}

		format.JSON(w, 200, Response{ID: int(id)})
	})

	r.Put("/tasks/{id}/position", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()

		id := NumberParam(r, "id")
		target := NumberFromForm(r.Form, "target", 0)
		parent := NumberFromForm(r.Form, "parent", -1)
		mode := r.Form.Get("mode")

		err := sendMoveQuery(id, mode, target, parent)
		if err != nil {
			format.Text(w, 500, err.Error())
			return
		}

		format.JSON(w, 200, Response{ID: id})
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
		id := NumberParam(r, "id")
		r.ParseForm()

		err := sendUpdateQuery("link", r.Form, id)
		if err != nil {
			format.Text(w, 500, err.Error())
			return
		}

		format.JSON(w, 200, Response{ID: id})
	})

	r.Delete("/links/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := NumberParam(r, "id")

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
		format.JSON(w, 200, Response{ID: int(id)})
	})

	r.Get("/resources", func(w http.ResponseWriter, r *http.Request) {
		data := make([]Resource, 0)

		err := conn.Select(&data, "SELECT resource.* FROM resource")
		if err != nil {
			format.Text(w, 500, err.Error())
			return
		}

		format.JSON(w, 200, data)
	})

	r.Get("/categories", func(w http.ResponseWriter, r *http.Request) {
		data := make([]Category, 0)

		err := conn.Select(&data, "SELECT category.* FROM category")
		if err != nil {
			format.Text(w, 500, err.Error())
			return
		}

		format.JSON(w, 200, data)
	})

	r.Get("/assignments", func(w http.ResponseWriter, r *http.Request) {
		data := make([]Assignment, 0)

		err := conn.Select(&data, "SELECT assignment.* FROM assignment")
		if err != nil {
			format.Text(w, 500, err.Error())
			return
		}

		format.JSON(w, 200, data)
	})

	r.Put("/assignments/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := NumberParam(r, "id")
		r.ParseForm()

		err := sendUpdateQuery("assignment", r.Form, id)
		if err != nil {
			format.Text(w, 500, err.Error())
			return
		}

		format.JSON(w, 200, Response{ID: id})
	})

	r.Delete("/assignments/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := NumberParam(r, "id")

		_, err := conn.Exec("DELETE FROM assignment WHERE id = ?", id)
		if err != nil {
			format.Text(w, 500, err.Error())
			return
		}

		format.JSON(w, 200, Response{ID: id})
	})

	r.Post("/assignments", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()

		res, err := sendInsertQuery("assignment", r.Form)
		if err != nil {
			format.Text(w, 500, err.Error())
			return
		}

		id, _ := res.LastInsertId()
		format.JSON(w, 200, Response{ID: int(id)})
	})

	log.Printf("Starting webserver at port " + Config.Port)
	http.ListenAndServe(Config.Port, r)
}

func setPosition(id int, mode string, parent int) error {
	if mode == "last" {
		var relatedPosition int
		row := conn.QueryRow("SELECT MAX(position)+1 from task WHERE parent = ?", parent)
		row.Scan(&relatedPosition)

		_, err := conn.Exec("UPDATE task SET position = ? WHERE id = ?", relatedPosition, id)
		return err
	} else if mode == "first" || mode == "" {
		_, err := conn.Exec("UPDATE task SET position = position + 1 WHERE parent = ? AND id <> ?", parent, id)
		return err
	}

	return errors.New("not supported position mode")
}

func sendMoveQuery(id int, mode string, target, parent int) error {
	var basePosition, baseParent int
	row := conn.QueryRow("SELECT parent, position from task WHERE id = ?", id)
	err := row.Scan(&baseParent, &basePosition)
	if err != nil {
		return err
	}

	var relatedPosition, relatedParent int
	relatedParent = parent
	if relatedParent == -1 {
		relatedParent = baseParent
	}

	if mode == "before" || mode == "after" {
		row := conn.QueryRow("SELECT parent, position from task WHERE id = ?", target)
		err = row.Scan(&relatedParent, &relatedPosition)
		if err != nil {
			return err
		}

		if mode == "after" {
			relatedPosition += 1
		}
	} else if mode == "last" {
		row := conn.QueryRow("SELECT MAX(position)+1 from task WHERE parent = ?", parent)
		row.Scan(&relatedPosition)
	}

	// source item removing may affect target index
	if relatedParent == baseParent && (mode == "last" || basePosition < relatedPosition) {
		relatedPosition -= 1
	}

	// already in place
	if relatedParent == baseParent && relatedPosition == basePosition {
		return nil
	}

	// removing from source order
	_, err = conn.Exec("UPDATE task SET position = position - 1 WHERE position > ? AND parent = ?", basePosition, baseParent)
	if err != nil {
		return err
	}
	// correct target order
	_, err = conn.Exec("UPDATE task SET position = position + 1 WHERE position >= ? AND parent = ?", relatedPosition, relatedParent)
	if err != nil {
		return err
	}
	// adding at target position
	_, err = conn.Exec("UPDATE task SET position = ?, parent = ? WHERE id = ?", relatedPosition, relatedParent, id)
	if err != nil {
		return err
	}

	return nil
}

// both task and link tables
var whitelistTask = []string{
	"text",
	"start_date",
	"duration",
	"parent",
	"progress",
	"opened",
	"details",
	"type",
}
var whitelistLink = []string{
	"source",
	"target",
	"type",
}
var whiteListAssignment = []string{
	"task",
	"resource",
	"value",
}

func getWhiteList(table string) []string {
	switch table {
	case "task":
		return whitelistTask
	case "link":
		return whitelistLink
	}
	return whiteListAssignment
}

func sendUpdateQuery(table string, form url.Values, id int) error {
	qs := "UPDATE " + table + " SET "
	params := make([]interface{}, 0)

	allowedFields := getWhiteList(table)
	for _, key := range allowedFields {
		value, ok := form[key]
		if ok {
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

func splitTask(parent int, form url.Values) (int, int, error) {
	var sibling int64

	// update parent - set it as type "split"
	_, err := conn.Exec("UPDATE task SET type = 'split', duration = GREATEST(duration, 1), progress = 0 WHERE id = ?", parent)
	if err != nil {
		return 0, 0, err
	}

	// add a clone-sibling if target parent doesn't already have at least 1 kid
	var hasKids bool
	row := conn.QueryRow("SELECT 1 from task WHERE parent = ? ORDER BY NULL LIMIT 1", parent)
	row.Scan(&hasKids)
	if !hasKids {
		res, err := conn.Exec("INSERT INTO task (text, start_date, type, duration, parent, progress, opened, details) SELECT text, start_date, 'task', GREATEST(duration,1), ?, GREATEST(progress,0), opened, details FROM task WHERE id = ?", parent, parent)
		if err != nil {
			return 0, 0, err
		}
		sibling, err = res.LastInsertId()
		if err != nil {
			return 0, 0, err
		}
	}

	res, err := sendInsertQuery("task", form)
	if err != nil {
		return 0, 0, err
	}
	id, _ := res.LastInsertId()

	return int(id), int(sibling), nil
}

func NumberParam(r *http.Request, key string) int {
	value := chi.URLParam(r, key)
	num, _ := strconv.Atoi(value)

	return num
}

func NumberFromForm(r url.Values, key string, defValue int) int {
	value := r.Get(key)
	if value == "" {
		return defValue
	}
	num, _ := strconv.Atoi(value)
	return num
}
