package main

import (
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

	//"github.com/pocketbase/pocketbase/tools/cron"

	"github.com/pocketbase/pocketbase/tools/mailer"
)

var mu sync.Mutex
var yourReplyToEmail = "help@`+ yourDomainVar + `"
var yourDomainVar = "note.suddsy.dev"

//var viewMu = &sync.Mutex{}

func main() {
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

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		// serves static files from the provided public dir (if exists)
		e.Router.GET("/*", apis.StaticDirectoryHandler(os.DirFS(publicDir), indexFallback))
		//scheduler := cron.New()
		//
		//// prints "Hello!" every 2 minutes
		//scheduler.MustAdd("hello", "0 */12 * * *", func() {
		//	arr := [9]string{"users", "pages", "imgs", "files", "subscriptions", "cookies", "reviews", "user_custom_settings", "bug_reports"}
		//	// Iterate through the array using range-based loop
		//	for _, value := range arr {
		//		res, err := app.Dao().DB().
		//			NewQuery(`DELETE FROM ` + value + `;`).
		//			Execute()
		//		if err != nil {
		//			return
		//		}
		//		log.Println(res)
		//
		//	}
		//
		//})
		//scheduler.Start()
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
		collection, err := app.Dao().FindCollectionByNameOrId("pages")
		if err != nil {
			return err
		}

		record := models.NewRecord(collection)

		// set individual fields
		// or bulk load with record.Load(map[string]any{...})
		record.Set("owner", e.Record.Id)
		record.Set("content", `{
			"time": 1696219628692,
			"blocks": [
			  {
				"id": "aV-p5oq-7s",
				"type": "paragraph",
				"data": {
				  "text": "👋 Welcome!"
				}
			  },
			  {
				"id": "NGCwdl_tCP",
				"type": "paragraph",
				"data": {
				  "text": "Here are the basics:"
				}
			  },
			  {
				"id": "xoPMiSDOBz",
				"type": "nestedList",
				"data": {
				  "style": "unordered",
				  "items": [
					{
					  "content": "Click anywhere and just start typing",
					  "items": []
					},
					{
					  "content": "Hit [<b>tab</b>] to see all the types of content you can add - photos, videos, sub pages, etc.",
					  "items": []
					},
					{
					  "content": "<mark class=\"cdx-marker\">Highlight </mark>any text, and use the menu that pops up to <b><i>style </i></b>your writing <b><mark class=\"cdx-marker\" style=\"background-color: rgb(204, 205, 245);\">however </mark></b>",
					  "items": []
					},
					{
					  "content": "See the ⋮⋮ to the left of this? Click it for more options with the block.",
					  "items": []
					},
					{
					  "content": "Click the + New Page button at the bottom of your sidebar to add a new page",
					  "items": []
					},
					{
					  "content": "Click Templates in your sidebar to get started with pre-built pages",
					  "items": []
					},
					{
					  "content": "Click the arrow in the side bar next to the page to view/create sub-pages",
					  "items": []
					}
				  ]
				}
			  },
			  {
				"id": "pFxqj9VLwS",
				"type": "paragraph",
				"data": {
				  "text": "👉&nbsp;Have a question 🤔? Send us a message!"
				}
			  },
			  {
				"id": "AC4vYAeixu",
				"type": "paragraph",
				"data": {
				  "text": "<i>Cover by:</i>&nbsp;<a href=\"https://unsplash.com/@vimal_saran\">unsplash.com/@vimal_saran</a>"
				}
			  }
			],
			"version": "2.27.2"
		  }`)
		record.Set("title", "Welcome")
		record.Set("icon", "1f44b.png")
		record.Set("unsplash", "https://images.unsplash.com/photo-1694365899936-850bc6c2b0f6?crop=entropy&cs=srgb&fm=jpg&ixid=M3w0NTQ4OTJ8MHwxfHJhbmRvbXx8fHx8fHx8fDE2OTYyMTk1MzJ8&ixlib=rb-4.0.3&q=85")

		if err := app.Dao().SaveRecord(record); err != nil {
			return err
		}
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
