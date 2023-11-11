package main

import (
	"encoding/json"
	"log"
	"net/mail"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/plugins/ghupdate"
	"github.com/pocketbase/pocketbase/plugins/jsvm"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
	"github.com/pocketbase/pocketbase/tools/cron"
	"github.com/pocketbase/pocketbase/tools/mailer"
	"github.com/spf13/cobra"
)

var mu sync.Mutex
var yourDomainVar = "note.suddsy.dev"

//var viewMu = &sync.Mutex{}

func main() {
	log.Println("test")
	//If using outside docker compose un comment these
	//err := godotenv.Load()
	//if err != nil {
	//	log.Fatal("Error loading .env file")
	//}
	autoReset := os.Getenv("AUTO_RESET")
	if autoReset == "true" {
		log.Println("AUTO RESET IS ACTIVE")
	}
	yourReplyToEmail := os.Getenv("REPLY_TO_EMAIL")
	if yourReplyToEmail == "" {
		yourReplyToEmail = "help@example.com"
	}
	app := pocketbase.New()

	// ---------------------------------------------------------------
	// Optional plugin flags:
	// ---------------------------------------------------------------

	var hooksDir string
	app.RootCmd.PersistentFlags().StringVar(
		&hooksDir,
		"hooksDir",
		"",
		"the directory with the JS app hooks",
	)

	var hooksWatch bool
	app.RootCmd.PersistentFlags().BoolVar(
		&hooksWatch,
		"hooksWatch",
		true,
		"auto restart the app on pb_hooks file change",
	)

	var hooksPool int
	app.RootCmd.PersistentFlags().IntVar(
		&hooksPool,
		"hooksPool",
		25,
		"the total prewarm goja.Runtime instances for the JS app hooks execution",
	)

	var migrationsDir string
	app.RootCmd.PersistentFlags().StringVar(
		&migrationsDir,
		"migrationsDir",
		"",
		"the directory with the user defined migrations",
	)

	var automigrate bool
	app.RootCmd.PersistentFlags().BoolVar(
		&automigrate,
		"automigrate",
		true,
		"enable/disable auto migrations",
	)

	var publicDir string
	app.RootCmd.PersistentFlags().StringVar(
		&publicDir,
		"publicDir",
		defaultPublicDir(),
		"the directory to serve static files",
	)

	var indexFallback bool
	app.RootCmd.PersistentFlags().BoolVar(
		&indexFallback,
		"indexFallback",
		true,
		"fallback the request to index.html on missing static path (eg. when pretty urls are used with SPA)",
	)

	var queryTimeout int
	app.RootCmd.PersistentFlags().IntVar(
		&queryTimeout,
		"queryTimeout",
		30,
		"the default SELECT queries timeout in seconds",
	)

	app.RootCmd.ParseFlags(os.Args[1:])

	// ---------------------------------------------------------------
	// Plugins and hooks:
	// ---------------------------------------------------------------

	// load jsvm (hooks and migrations)
	jsvm.MustRegister(app, jsvm.Config{
		MigrationsDir: migrationsDir,
		HooksDir:      hooksDir,
		HooksWatch:    hooksWatch,
		HooksPoolSize: hooksPool,
	})

	// migrate command (with js templates)
	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		TemplateLang: migratecmd.TemplateLangJS,
		Automigrate:  automigrate,
		Dir:          migrationsDir,
	})

	// GitHub selfupdate
	ghupdate.MustRegister(app, app.RootCmd, ghupdate.Config{})

	app.OnAfterBootstrap().Add(func(e *core.BootstrapEvent) error {
		app.Dao().ModelQueryTimeout = time.Duration(queryTimeout) * time.Second
		return nil
	})

	createWelcomePage := func(user string) error {
		collection, err := app.Dao().FindCollectionByNameOrId("pages")
		if err != nil {
			return err
		}

		record := models.NewRecord(collection)

		// set individual fields
		// or bulk load with record.Load(map[string]any{...})
		executable, err := os.Executable()
		if err != nil {
			log.Fatal("Error getting executable path:", err)
		}

		filePath := filepath.Join(filepath.Dir(executable), "preview_page.json")
		data, err := os.ReadFile(filePath)
		if err != nil {
			log.Println(err)
			return err
		}
		type Page struct {
			Content  json.RawMessage `json:"content"`
			Shared   bool            `json:"shared"`
			Id       string          `json:"id"`
			Title    string          `json:"title"`
			Icon     string          `json:"icon"`
			Unsplash string          `json:"unsplash"`
		}
		var page Page
		err = json.Unmarshal(data, &page)
		if err != nil {
			log.Println(err)
			return err
		}
		if user != "" {
			record.Set("owner", user)
		}
		record.Set("content", page.Content)
		record.Set("shared", page.Shared)
		record.Set("id", page.Id)
		record.Set("title", page.Title)
		record.Set("icon", page.Icon)
		record.Set("unsplash", page.Unsplash)

		if err := app.Dao().SaveRecord(record); err != nil {
			log.Println(err)
			return err
		}
		return nil
	}

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		// serves static files from the provided public dir (if exists)
		e.Router.GET("/*", apis.StaticDirectoryHandler(os.DirFS(publicDir), indexFallback))

		//Auto drop all tables
		scheduler := cron.New()
		if autoReset == "true" {
			scheduler.MustAdd("hello", "0 */12 * * *", func() {
				arr := [9]string{"users", "pages", "imgs", "files"}
				// Iterate through the array using range-based loop
				for _, value := range arr {
					_, err := app.Dao().DB().
						NewQuery(`DELETE FROM ` + value + `;`).
						Execute()
					if err != nil {
						return
					}
					//log.Println(res)
					createWelcomePage("")
				}
			})
		}
		scheduler.Start()
		createWelcomePage("")
		return nil
	})

	sendEmail := func(email string, username string, ip string) {
		//muLogin.Lock()         // Lock the mutex before accessing the shared resource (if necessary)
		//defer muLogin.Unlock() // Ensure the mutex is unlocked even if a panic occurs

		message := &mailer.Message{
			From: mail.Address{
				Address: app.Settings().Meta.SenderAddress,
				Name:    app.Settings().Meta.SenderName,
			},
			To:      []mail.Address{{Address: email}},
			Subject: `New login from  ` + ip,
			HTML:    `<p>Hello, ` + username + `</p><p>This is an email to let you know you have had a login on your account from ` + ip + `</p><p>Click on the button below to change your password if this was not you.</p><p><a style="display: inline-block; vertical-align: top; border: 0; color: #fff!important; background: #16161a!important; text-decoration: none!important; line-height: 40px; width: auto; min-width: 150px; text-align: center; padding: 0 20px; margin: 5px 0; font-family: Source Sans Pro, sans-serif, emoji; font-size: 14px; font-weight: bold; border-radius: 6px; box-sizing: border-box;" href="https://` + yourDomainVar + `/auth/pwdreset" target="_blank" rel="noopener">Change password</a></p><p>Thanks,<br/>` + yourDomainVar + ` team</p>`, // bcc, cc, attachments and custom headers are also supported...
			Headers: map[string]string{
				"Reply-To": yourReplyToEmail,
			},
		}
		app.NewMailClient().Send(message)
	}

	app.OnRecordAfterCreateRequest("users").Add(func(e *core.RecordCreateEvent) error {
		//log.Println(e.Record)

		createWelcomePage(e.Record.Id)
		return nil
	})

	app.OnRecordAfterAuthWithPasswordRequest("users").Add(func(e *core.RecordAuthWithPasswordEvent) error {
		usrEmail := e.Record.Email()
		usrUsername := e.Record.Username()
		loginIp := e.HttpContext.RealIP()
		if e.Record.Verified() {
			go sendEmail(usrEmail, usrUsername, loginIp)
		}

		return nil
	})

	app.OnRecordAfterAuthWithOAuth2Request("users").Add(func(e *core.RecordAuthWithOAuth2Event) error {
		usrEmail := e.Record.Email()
		usrUsername := e.Record.Username()
		loginIp := e.HttpContext.RealIP()
		if e.Record.Verified() {
			go sendEmail(usrEmail, usrUsername, loginIp)
		}

		return nil
	})

	app.OnRecordBeforeCreateRequest("imgs", "files").Add(func(e *core.RecordCreateEvent) error {
		admin, _ := e.HttpContext.Get(apis.ContextAdminKey).(*models.Admin)
		if admin != nil {
			return nil
		}
		authRecord, _ := e.HttpContext.Get(apis.ContextAuthRecordKey).(*models.Record)
		if authRecord == nil {
			return apis.NewForbiddenError("Only auth records can access this endpoint", nil)
		}
		if authRecord.CleanCopy().GetBool("admin") {
			return nil
		}
		recordimg, err1 := app.Dao().FindRecordById("Total_img_per_user", authRecord.Id)
		recordfile, err2 := app.Dao().FindRecordById("total_files_per_user", authRecord.Id)

		if err1 != nil && err2 != nil {
			return nil // Both records couldn't be fetched, handle the error accordingly
		}

		var totalSize float64

		if err1 == nil {
			totalSize += recordimg.GetFloat("total_size")
		}

		if err2 == nil {
			totalSize += recordfile.GetFloat("total_size")
		}

		//log.Println(totalSize)

		if totalSize > 10485760 {
			return apis.NewForbiddenError("You have exceeded the total size of embeds/images allowed for your account.", nil)
		}

		// Continue with your code logic if the total size is within the limit

		return nil
	})

	//Run after the record is created
	app.OnRecordAfterCreateRequest("imgs", "files").Add(func(e *core.RecordCreateEvent) error {
		if e.UploadedFiles == nil {
			return nil
		}
		//Create a virtual filesystem (if s3 else just normal fs)
		files, err := app.NewFilesystem()
		if err != nil {
			return apis.NewBadRequestError("Data error fs", nil)
		}
		defer files.Close()
		var size int64

		//Works with & without S3
		//Get the file
		file, err2 := files.GetFile(e.Collection.Id + "/" + e.Record.Id + "/" + e.Record.Get("file_data").(string))
		if err2 != nil {
			return apis.NewBadRequestError("Data error fs find", nil)
		}
		//Get the file size
		size = file.Size()
		//Close the file reader
		files.Close()

		//Find the record for that file
		record, err := app.Dao().FindRecordById(e.Collection.Name, e.Record.Id)
		if err != nil {
			return apis.NewBadRequestError("Data error col", nil)
		}
		//Set the size for that record to the files size (bytes)
		record.Set("size", size)
		//Save the record
		if err := app.Dao().SaveRecord(record); err != nil {
			return err
		}
		return nil
	})

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.GET("/ping", func(c echo.Context) error {
			//Get current user from a auth record
			authRecord, _ := c.Get(apis.ContextAuthRecordKey).(*models.Record)
			if authRecord == nil {
				return apis.NewForbiddenError("Only auth records can access this endpoint", nil)
			}
			//Find the users authrecord
			//This has authRecord.Collection().Name so that it will work with any auth collection
			record, _ := app.Dao().FindRecordById(authRecord.Collection().Name, authRecord.Id)
			record.Set("last_active", time.Now().UTC())

			if err := app.Dao().SaveRecord(record); err != nil {
				return err
			}

			return c.String(200, "Received ping :)")
		})

		return nil
	})

	app.RootCmd.AddCommand(&cobra.Command{
		Use: "deploy",
		Run: func(cmd *cobra.Command, args []string) {
			updater()
		},
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}

// the default pb_public dir location is relative to the executable
func defaultPublicDir() string {
	if strings.HasPrefix(os.Args[0], os.TempDir()) {
		// most likely ran with go run
		return "./pb_public"
	}

	return filepath.Join(os.Args[0], "../pb_public")
}
