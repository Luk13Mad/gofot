package main

import (
	"database/sql"
	"gofot/routes"
	"gofot/util"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var Templates = loadTemplates()
var db *sql.DB
var DATABASE string //= "postgres://POSTGRES_USR:POSTGRES_PWD@localhost:1234/POSTGRES_DB?sslmode=disable"

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags)

	logger.Println("Server is starting...")

	err := godotenv.Load()
	if err != nil {
		logger.Fatal("Error loading .env file")
	}

	PORT := os.Getenv("PORT")
	postgres_usr, postgres_pwd, postgres_db, postgres_port := os.Getenv("POSTGRES_USR"), os.Getenv("POSTGRES_PWD"), os.Getenv("POSTGRES_DB"), os.Getenv("POSTGRES_PORT")
	DATABASE = "postgres://" + postgres_usr + ":" + postgres_pwd + "@localhost:" + postgres_port + "/" + postgres_db + "?sslmode=disable"

	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("./static/")) //fileserver for static files

	db, err := sql.Open("postgres", DATABASE) //open connection to DB and fail on error
	if err != nil {
		logger.Fatal(err)
	}
	logger.Println("Connected to DB...")

	max_open_conns, err := strconv.Atoi(os.Getenv("MAXOPENCONNS"))
	if err != nil {
		logger.Fatal(err)
	}
	max_idle_conns, err := strconv.Atoi(os.Getenv("MAXIDLECONNS"))
	if err != nil {
		logger.Fatal(err)
	}
	max_conn_lifetime, err := strconv.Atoi(os.Getenv("MAXCONNLIFETIME"))
	if err != nil {
		logger.Fatal(err)
	}
	db.SetMaxOpenConns(max_open_conns) //set DB connection parameters
	db.SetMaxIdleConns(max_idle_conns)
	db.SetConnMaxLifetime(time.Duration(max_conn_lifetime) * time.Minute)

	if err = db.Ping(); err != nil { //ping DB and fail on error
		logger.Fatal(err)
	}

	defer db.Close() //defer closing connection to DB

	//handle wrapper used to carry templates and DB conecction
	//also handles logger and tracing
	var mhw = routes.NewHandleWrapper(Templates, db,
		logger, func() string { return strconv.FormatInt(time.Now().UnixNano(), 36) })

	mux.HandleFunc("/", mhw.Index)                                    //only GET
	mux.HandleFunc("/2D", mhw.Process2D)                              //GET and POST
	mux.HandleFunc("/2D_form", mhw.Process2DForm)                     //only POST
	mux.HandleFunc("/3D", mhw.Process3D)                              //GET and POST
	mux.HandleFunc("/3D_form", mhw.Process3DForm)                     //only POST
	mux.HandleFunc("/faq", mhw.ProcessFAQ)                            //only GET
	mux.HandleFunc("/contact", mhw.ProcessContact)                    //only GET
	mux.HandleFunc("/about", mhw.ProcessAbout)                        //only GET
	mux.HandleFunc("/download_2D_example", mhw.ProcessDwnl2DExamples) //only GET
	mux.HandleFunc("/download_3D_example", mhw.ProcessDwnl3DExamples) //only GET
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	server := &http.Server{
		Addr:         ":" + PORT,
		Handler:      (util.Middlewares{mhw.Logging, mhw.Tracing}).Apply(mux), //successively apply all middlewares staring from the left
		ErrorLog:     logger,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		logger.Fatalf("Could not listen on %q: %s\n", PORT, err)
	}
}

func loadTemplates() map[string]*template.Template {
	tmpl := make(map[string]*template.Template)
	tmpl["index.html"] = template.Must(template.ParseFiles("./html/index.html", "./html/base.html"))
	tmpl["2D_genes.html"] = template.Must(template.ParseFiles("./html/2D_genes.html", "./html/base.html"))
	tmpl["3D_genes.html"] = template.Must(template.ParseFiles("./html/3D_genes.html", "./html/base.html"))
	tmpl["404.html"] = template.Must(template.ParseFiles("./html/404.html", "./html/base.html"))
	tmpl["about.html"] = template.Must(template.ParseFiles("./html/about.html", "./html/base.html"))
	tmpl["contact.html"] = template.Must(template.ParseFiles("./html/contact.html", "./html/base.html"))
	tmpl["error_message.html"] = template.Must(template.ParseFiles("./html/error_message.html", "./html/base.html"))
	tmpl["faq.html"] = template.Must(template.ParseFiles("./html/faq.html", "./html/base.html"))
	return tmpl
}
