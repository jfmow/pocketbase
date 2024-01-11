### To compile to rpi 32bit from Windows:

- Run the make.py script and select `n` for upload to the server, ensuring you have an already running instance.

- This will build a `base` in the current directory, which you can copy over to your server to run.

**Alert:** Add the following environment variables:

---

- **REPLY_TO_EMAIL="hi@yourdomain.com"**
  - The email that users can reply to your emails from.

- **AUTO_RESET="false"**
  - Prevents the database from dropping all tables every 12 hours.

- **updateApiKey="hTu/9tNc4IpsRa9kfuIgSWS1LhCZTr6fu/E5/uwt4bPYqYX0YEkaJJxJMA=="**
  - A random base64 string used to authenticate the base instance with the update base.

- **updateApiUrl="https://yourdomain.com/update/latest?auth="**
  - The URL of the base containing the updated base, selecting the most recent. Can be the current base instance. **URL must be in this format**.

- **RESEND_API="re_xyz..."**
  - Your resend API key to send emails.

- **RESEND_EMAIL="hi@yourdomain.com"**
  - The email address from which to send emails.
