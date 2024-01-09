package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"net/mail"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/plugins/ghupdate"
	"github.com/pocketbase/pocketbase/plugins/jsvm"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
	"github.com/pocketbase/pocketbase/tools/cron"
	"github.com/pocketbase/pocketbase/tools/mailer"
	"github.com/pocketbase/pocketbase/tools/security"
	"github.com/spf13/cobra"
)

// var viewMu = &sync.Mutex{}
func generateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes)[:length], nil
}
func isValidTokenFormat(token string) bool {
	// Define a regular expression for your desired format
	// For example, assuming a 24-character hexadecimal string
	validFormat := regexp.MustCompile(`^[0-9a-fA-F]{24}$`)
	return validFormat.MatchString(token)
}
func isValidEmailFormat(email string) bool {
	validFormat := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return validFormat.MatchString(email)
}

func main() {
	//If using outside docker compose un comment these
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}
	//gitHub()
	autoReset := os.Getenv("AUTO_RESET")
	if autoReset == "true" {
		log.Println("AUTO RESET IS ACTIVE")
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
		if user == "" {
			record.Set("id", page.Id)
		}
		record.Set("title", page.Title)
		record.Set("icon", page.Icon)
		record.Set("unsplash", page.Unsplash)

		if err := app.Dao().SaveRecord(record); err != nil {
			log.Println(err)
			return err
		}
		return nil
	}
	createDefaultUserFlagsRecord := func(userId string, ssoEnabled bool) error {
		collection, err := app.Dao().FindCollectionByNameOrId("user_flags")
		if err != nil {
			return err
		}

		record := models.NewRecord(collection)

		// set individual fields
		// or bulk load with record.Load(map[string]any{...})
		record.Set("user", userId)
		record.Set("quota", 10485760)
		if ssoEnabled {
			record.Set("sso", true)
		}

		if err := app.Dao().SaveRecord(record); err != nil {
			return err
		}
		return nil
	}

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		// serves static files from the provided public dir (if exists)
		e.Router.GET("/*", apis.StaticDirectoryHandler(os.DirFS(publicDir), indexFallback))
		//TODO: make this use a token randomly generated by giving the old token and getting a new one to use for this
		e.Router.GET("/update/latest", func(c echo.Context) error {
			updateApiKey := c.QueryParam("auth")
			updateApiKeyEnv := os.Getenv("updateApiKey")
			//log.Println(updateApiKeyEnv, updateApiKey)
			if updateApiKey != updateApiKeyEnv {
				return c.JSON(http.StatusForbidden, "")
			}
			collId, err := app.Dao().FindCollectionByNameOrId("pocketbases")
			if err != nil {
				return c.JSON(http.StatusInternalServerError, "")
			}
			type Update struct {
				Id   string `db:"id" json:"id"`
				Base string `db:"base" json:"base"`
			}
			result := Update{}

			app.Dao().DB().
				Select("pocketbases.*").
				From("pocketbases").
				OrderBy("updated DESC").One(&result)

			newFs, err := app.NewFilesystem()
			if err != nil {
				return c.JSON(http.StatusInternalServerError, "")
			}
			return newFs.Serve(c.Response().Writer, c.Request(), collId.Id+"/"+result.Id+"/"+result.Base, "base")

			//c.File()

		} /* optional middlewares */)
		e.Router.GET("/analytics", func(c echo.Context) error {
			id := c.QueryParam("id")
			url := c.QueryParam("url")
			action := c.QueryParam("action")
			ip := c.RealIP()
			region := strings.Split(c.Request().Header.Get("accept-language"), ",")[0]
			if id == "" || url == "" || ip == "" || region == "" {
				return c.String(400, "Missing required data")
			}

			collection, err := app.Dao().FindCollectionByNameOrId("analytics")
			if err != nil {
				return c.NoContent(500)
			}

			record := models.NewRecord(collection)

			record.Set("fingerprint", id)
			record.Set("url", url)
			record.Set("ip", ip)
			record.Set("action", action)
			record.Set("region", region)

			if err := app.Dao().SaveRecord(record); err != nil {
				return c.NoContent(500)
			}
			return c.String(200, "Success")
		})
		e.Router.GET("/analytics/reset_table", func(c echo.Context) error {
			_, err := app.Dao().DB().
				NewQuery(`DELETE FROM ` + "analytics" + `;`).
				Execute()
			if err != nil {
				return c.String(500, "Failed to reset analytics table")
			}
			return c.String(200, "Complete")
		}, apis.RequireAdminAuth())
		e.Router.POST("/api/auth/sso/signup", func(c echo.Context) error {
			email := c.QueryParam("email")
			username := c.QueryParam("username")
			if email == "" || username == "" || !isValidEmailFormat(email) {
				return apis.NewForbiddenError("Missing required data/Incorrect formatting of required data", nil)
			}
			collection, err := app.Dao().FindCollectionByNameOrId("users")
			if err != nil {
				return apis.NewApiError(500, "Failed to search", nil)
			}

			user, err := app.Dao().FindAuthRecordByEmail("users", email)
			if user != nil && err != nil {
				return apis.NewBadRequestError("Failed to create account", nil)
			}

			//Generate token key for account (REQUIRED)

			randomKey := security.RandomString(50)

			record := models.NewRecord(collection)
			record.Set("email", email)
			record.Set("username", username)
			record.Set("tokenKey", randomKey)
			if err := app.Dao().SaveRecord(record); err != nil {
				log.Println(err)
				return apis.NewBadRequestError("Failed to create account", nil)
			}
			record2, err2 := app.Dao().FindFirstRecordByData("users", "email", email)
			if err2 != nil {
				return apis.NewBadRequestError("Failed to login account", nil)
			}
			info := apis.RequestInfo(c)
			canAccess, err := app.Dao().CanAccessRecord(record2, info, record2.Collection().CreateRule)
			if !canAccess {
				if err := app.Dao().DeleteRecord(record2); err != nil {
					return apis.NewForbiddenError("", err)
				}
				return apis.NewForbiddenError("", err)
			}
			go createWelcomePage(record2.Id)
			if err := createDefaultUserFlagsRecord(record2.Id, true); err != nil {
				return apis.NewApiError(500, "Flag creation failed", nil)
			}

			return apis.RecordAuthResponse(app, c, record2, nil)

		}, apis.ActivityLogger(app))
		e.Router.POST("/api/auth/sso/login", func(c echo.Context) error {
			email := c.QueryParam("email")
			token := c.QueryParam("token")
			if email == "" || token == "" || !isValidEmailFormat(email) || !isValidTokenFormat(token) {
				return apis.NewForbiddenError("Missing required data/Incorrect formatting of required data", nil)
			}
			//Find token in db
			record, err := app.Dao().FindFirstRecordByData("sso_tokens", "token", token)
			if err != nil {
				return apis.NewBadRequestError("Invalid credentials", nil)
			}
			//Check if its still valid
			if record.GetDateTime("valid_until").Time().Before(time.Now()) {
				if err := app.Dao().DeleteRecord(record); err != nil {
					return err
				}
				return apis.NewBadRequestError("Invalid credentials", nil)
			}
			//Compare sent email to tokens email
			if record.Get("email") != email {
				return apis.NewBadRequestError("Invalid credentials", nil)
			}
			//Login
			record2, err2 := app.Dao().FindFirstRecordByData("users", "email", email)
			if err2 != nil || record2.Email() != email {
				// return generic 400 error to prevent phones enumeration
				return apis.NewBadRequestError("Invalid credentials", nil)
			}
			if err := app.Dao().DeleteRecord(record); err != nil {
				return apis.NewBadRequestError("Invalid credentials", nil)
			}
			return apis.RecordAuthResponse(app, c, record2, nil)
		}, apis.ActivityLogger(app))
		e.Router.POST("/api/auth/sso", func(c echo.Context) error {
			email := c.QueryParam("email")
			linkUrl := c.QueryParam("linkUrl")
			currentTime := time.Now().UTC()
			// Add 5 minutes to the current time
			newTime := currentTime.Add(5 * time.Minute)
			randomString, err := generateRandomString(24)
			if err != nil {
				return apis.NewBadRequestError("Unable to generate token!", nil)
			}

			//Check that email exists and is valid format
			if email == "" || !isValidEmailFormat(email) {
				return apis.NewForbiddenError("Missing required data/Incorrect formatting of required data", nil)
			}

			//Check if user exists
			AuthRecord, err := app.Dao().FindFirstRecordByData("users", "email", email)
			if AuthRecord == nil || err != nil {
				return apis.NewBadRequestError("No user found with that email!", nil)
			}

			userFlagsRecord, err := app.Dao().FindFirstRecordByData("user_flags", "user", AuthRecord.Id)
			if err != nil {
				return apis.NewBadRequestError("Record error", nil)
			}
			if !userFlagsRecord.GetBool("sso") {
				return apis.NewUnauthorizedError("Auth method not enabled", nil)
			}

			//Check if user already has a pending/forgotten token.
			recordA, _ := app.Dao().FindFirstRecordByData("sso_tokens", "email", email)
			if recordA != nil {
				//If there is one left, update its values instead of deleting.
				recordA.Set("token", randomString)
				recordA.Set("valid_until", newTime)
				if err := app.Dao().SaveRecord(recordA); err != nil {
					return apis.NewBadRequestError("Unable to create token", nil)
				}
			} else {
				//There isn't one, so create a new db entry
				collection, err := app.Dao().FindCollectionByNameOrId("sso_tokens")
				if err != nil {
					return apis.NewBadRequestError("Unable to create token", nil)
				}

				record := models.NewRecord(collection)
				record.Set("email", email)
				record.Set("token", randomString)
				record.Set("valid_until", newTime)

				if err := app.Dao().SaveRecord(record); err != nil {
					return apis.NewBadRequestError("Unable to create token", nil)
				}
			}

			htmlString := `
			<table width="100%" border="0" cellspacing="0" cellpadding="0" style="background:#f4f4f4; padding: 48px; margin: 0;">
				<tr>
					<td align="center">
						<div
							style="max-width: 460px; height: fit-content; padding: 48px; color: #111111; background: #ffffff; word-wrap: break-word; line-height: 1.6; border-radius: 0.5rem; box-shadow: 2.8px 2.8px 2.2px -19px rgba(0, 0, 0, 0.07), 6.7px 6.7px 5.3px -19px rgba(0, 0, 0, 0.05), 12.5px 12.5px 10px -19px rgba(0, 0, 0, 0.042), 22.3px 22.3px 17.9px -19px rgba(0, 0, 0, 0.035), 41.8px 41.8px 33.4px -19px rgba(0, 0, 0, 0.028), 100px 100px 80px -19px rgba(0, 0, 0, 0.02);">
							<img src="https://p.suddsy.dev/Favicon.png" alt="Company Logo"
								style="width: 40px; height: 40px; margin: 10px; margin-left: 0;">
							<h1 style="font-size: 34px;">Sign-in with SSO</h1>
							<p style="font-weight: inherit; line-height: 1.6; font-size: 18px; margin: 0 0 12px; padding: 0;">
								To login copy this token and paste it in the login screen</p>
							<p style="font-weight: inherit; line-height: 1.6; font-size: 18px; margin: 0 0 12px; padding: 0;">{TOKEN HERE}</p>
							<p style="font-weight: inherit; line-height: 1.6; font-size: 18px; margin: 0 0 12px; padding: 0;">Or click this magic link: <a  target="_blank" rel="noopener" style="color: #999999; text-decoration: underline; cursor: pointer;" href='{APPURL HERE}/auth/login?ssoToken={TOKEN HERE}&ssoEmail={USER EMAIL HERE}'>login</a></p>
						</div>
					</td>
				</tr>
			</table>
			`

			// Token to be inserted
			token := randomString

			// Replace {TOKEN HERE} with the actual token
			modifiedHTML1 := strings.Replace(htmlString, "{TOKEN HERE}", token, 2)
			modifiedHTML2 := strings.Replace(modifiedHTML1, "{APPURL HERE}", linkUrl, 1)
			modifiedHTML := strings.Replace(modifiedHTML2, "{USER EMAIL HERE}", email, 1)

			message := &mailer.Message{
				From: mail.Address{
					Address: app.Settings().Meta.SenderAddress,
					Name:    app.Settings().Meta.SenderName,
				},
				To:      []mail.Address{{Address: email}},
				Subject: "SSO Token",
				HTML:    modifiedHTML,
				// bcc, cc, attachments and custom headers are also supported...
			}

			go app.NewMailClient().Send(message)
			return nil
		})
		e.Router.POST("/api/auth/sso/toggle", func(c echo.Context) error {
			authRecord, _ := c.Get(apis.ContextAuthRecordKey).(*models.Record)
			newPassword := c.QueryParam("np")
			isGuest := authRecord == nil
			if isGuest {
				return apis.NewForbiddenError("Must be signed in to preform this action", nil)
			}
			record, err := app.Dao().FindFirstRecordByData("user_flags", "user", authRecord.Id)
			if err != nil {
				return apis.NewBadRequestError("Problem while checking status", nil)
			}

			if authRecord.PasswordHash() == "" && newPassword == "" && record.GetBool("sso") {
				return apis.NewBadRequestError("You must set a password before disabling SSO", nil)
			}
			if newPassword != "" {
				if err := authRecord.SetPassword(newPassword); err != nil {
					return apis.NewBadRequestError("Missing or invalid data", nil)
				}
				if !authRecord.ValidatePassword(newPassword) {
					return apis.NewApiError(500, "Unable to validate new password", nil)
				}
				if err := app.Dao().SaveRecord(authRecord); err != nil {
					return apis.NewApiError(500, "Unable to set new password", nil)
				}
			}

			record.Set("sso", !record.GetBool("sso"))

			if err := app.Dao().SaveRecord(record); err != nil {
				return apis.NewApiError(500, "Unable to update sso state", nil)
			}
			return nil
		}, apis.ActivityLogger(app))
		//Auto drop all tables
		scheduler := cron.New()
		if autoReset == "true" {
			scheduler.MustAdd("autoDelete", "0 */12 * * *", func() {
				arr := [9]string{"users", "pages", "imgs", "files", "user_flags"}
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
		scheduler.MustAdd("ssoTokenClear", "*/5 * * * *", func() {
			currentTime := time.Now().UTC()
			_, err := app.Dao().DB().
				NewQuery("DELETE FROM sso_tokens WHERE valid_until < {:date}").
				Bind(dbx.Params{
					"date": currentTime,
				}).Execute()
			if err != nil {
				// Handle the error
				log.Println(err)
				return
			}
		})
		scheduler.Start()
		createWelcomePage("")
		return nil
	})

	app.OnRecordAfterCreateRequest("users").Add(func(e *core.RecordCreateEvent) error {
		//log.Println(e.Record)
		createWelcomePage(e.Record.Id)
		if err := createDefaultUserFlagsRecord(e.Record.Id, false); err != nil {
			return apis.NewApiError(500, "Flag creation failed", nil)
		}
		return nil
	})

	app.OnRecordBeforeCreateRequest("imgs", "files").Add(func(e *core.RecordCreateEvent) error {
		admin, _ := e.HttpContext.Get(apis.ContextAdminKey).(*models.Admin)
		if admin != nil {
			return nil
		}
		authRecord, _ := e.HttpContext.Get(apis.ContextAuthRecordKey).(*models.Record)
		flagsCollection, err := app.Dao().FindCollectionByNameOrId("user_flags")
		if err != nil || flagsCollection == nil {
			return nil
		}
		userFlags, err := app.Dao().FindFirstRecordByData(flagsCollection.Id, "user", authRecord.Id)
		if err != nil {
			return apis.NewBadRequestError("Failed to validate user quota. Please contact support if this issue persists.", nil)
		}
		//log.Println(userFlags)
		if userFlags.GetBool("admin") {
			return nil
		}
		if authRecord == nil {
			return apis.NewForbiddenError("Only auth records can access this endpoint", nil)
		}
		if authRecord.CleanCopy().GetBool("admin") {
			return nil
		}

		var totalColl int
		var totalSize float64

		recordimg, err := app.Dao().FindRecordById("Total_img_per_user", authRecord.Id)

		if err != nil {
			totalColl++
		} else {
			totalSize += recordimg.GetFloat("total_size")
		}

		recordfile, err := app.Dao().FindRecordById("total_files_per_user", authRecord.Id)

		if err != nil {
			totalColl++
		} else {
			totalSize += recordfile.GetFloat("total_size")
		}

		if totalColl >= 2 {
			return nil
		}

		if totalSize >= userFlags.GetFloat("quota") {
			return apis.NewForbiddenError("You have exceeded the total size of embeds/images allowed for your account.", nil)
		}

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
		Use: "updateme",
		Run: func(cmd *cobra.Command, args []string) {
			installUpdate()
		},
	})

	KillTheOldExe()

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
