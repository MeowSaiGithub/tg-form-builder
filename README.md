# 📝 Dynamic Form Telegram Bot

A powerful Telegram bot that dynamically generates forms based on a provided JSON configuration. It supports storing responses in MySQL, PostgreSQL, or MongoDB and can send data via a webhook.

## ✨ Features

- 📜 **Dynamic Form Generation** – Define forms via JSON.
- 💾 **Database Support** – Store responses in **MySQL, PostgreSQL, SQLite or MongoDB**.
- 🔄 **Webhook Integration** – Send data to external services.
- 📸 **Media Support** – Forms can include **photos, videos, and documents**.
- ✅ **Validation & Preprocessing** – Supports input validation and required fields.
- 🎨 **Customizable Buttons & Messages** – Forms can include inline buttons for user interaction and custom messages can be set.
- 🛠️ **Debug Mode & Memory Load** – Helps with performance tuning and debugging.
- 🔧 **Custom Executables** – Custom Executables.

## 🚧 Limitation 

- Only Singular flow is currently supported (which meant the form cannot dynamically change the flow).
- Previously shown buttons and selections can be clicked again. ( will remove/disable already selected options in the future)

## 🎬 Demo

- See the bot in action! Watch our short demo video: [Demo Video](https://youtu.be/sa20Ms3TtRs)

## 🚀 Getting Start

Download the bot from [Latest Release](https://github.com/MeowSaiGithub/tg-form-builder/releases/latest). You can download with multiple options -

- `all` tag include all databases `mysql`, `postgres`, and `mongo`
- `mongo` tag include only `mongo` database
- `postgres` tag include only `postgres` database
- `mysql` tag include only `mysql` database
- `sqlite` tag include only `sqlite` database
- `none` tag include only `blank` database


### 💻 Commands
### To validate the `format.json` file
```shell
  gotgbot validate -f format.json 
```

### To migrate the database schema
```shell
  gotgbot migrate -f format.json -c config.yaml
```

### To run the bot
```shell
  gotgbot start -f format.json -c config.yaml
```

## 🛠️ Configuration (`config.yaml`)

Modify `config.yaml` to customize the bot:

```yaml
debug_mode: true  # Enables detailed logging
enable_memory_load: true  # Loads form data into memory for faster access
memory_limit_mb: 1024  # Sets a memory limit (in MB)

bot:
  token: "YOUR_TELEGRAM_BOT_TOKEN" # Your Telegram bot token

webhook:
  enabled: false # Enable webhook
  url: "http://your-webhook-url/api" # Webhook URL
  workers_count: 5 # Number of webhook workers
  queue_size: 10 # Webhook queue size
  auth: # Webhook authentication
    type: "basic" # Choose from "none", "basic", and "bearer"
    token: "bearer-token" # Webhook authentication token for "bearer"
    username: "username" # Webhook authentication username for "basic"
    password: "password" # Webhook authentication password for "basic"

database:
  enable: true  # Enables database support
  use_adaptor: "sqlite"  # Choose from mysql, postgres, sqlite, or mongo
  mysql:
    #    dsn: ""  # MySQL DSN
    username: "username"  # MySQL username
    password: "password"  # MySQL password
    host: "localhost"  # MySQL host
    port: 3306  # MySQL port
    database: "db-name"  # MySQL database name
  mongo:
    uri: "mongodb://localhost:27017"
    #    addresses:
    #      - "localhost:27017"
    #    database: "db-name"
    #    username: "username"
    #    password: "password"
    #    replica-set: "rs0"
    #    auth-mechanism: "SCRAM-SHA-1"

  postgres:
    #    dsn: "postgresql://username:password@localhost:5432/dbname?sslmode=disable"
    username: "username"
    password: "password"
    host: "localhost"
    port: 5432
    database: "database-name"

  sqlite:
    dsn: "tf.db"
```
# 📄 JSON Form Format Explanation

This document provides a detailed explanation of how to define forms using JSON format for the Telegram bot.

---

## 📝 Example JSON Form

Below is a sample JSON form configuration for a **Customer Satisfaction Survey**.

```json
{
  "form_name": "Customer Satisfaction Survey",
  "table_name": "customer_satisfaction_responses",
  "review_enabled": true,
  "db": "mysql",
  "submit_message": "Thank you for your feedback!",
  "fields": [
    {
      "name": "name",
      "label": "Your Name",
      "type": "text",
      "db_type": "VARCHAR(100)",
      "required": true,
      "skippable": false,
      "description": "Please enter your full name.",
      "formatting": "HTML",
      "buttons": []
    },
    {
      "name": "email",
      "label": "Your Email",
      "type": "email",
      "db_type": "VARCHAR(255)",
      "required": true,
      "skippable": false,
      "description": "Please enter a valid email address.",
      "formatting": "HTML",
      "buttons": []
    },
    {
      "name": "age",
      "label": "Your Age",
      "type": "number",
      "db_type": "INT",
      "required": true,
      "skippable": false,
      "description": "Please enter your age (must be between 18 and 100).",
      "formatting": "Markdown",
      "validation": {
        "min": 18,
        "max": 100
      },
      "buttons": []
    },
    {
      "name": "feedback",
      "label": "Your Feedback",
      "type": "text",
      "db_type": "TEXT",
      "required": false,
      "skippable": true,
      "description": "We value your feedback. Please share your thoughts.",
      "formatting": "HTML",
      "buttons": []
    },
    {
      "name": "rating",
      "label": "Rate our service",
      "type": "select",
      "options": ["1", "2", "3", "4", "5"],
      "db_type": "VARCHAR(5)",
      "required": true,
      "skippable": false,
      "description": "Please rate our service from 1 to 5.",
      "formatting": "Markdown",
      "buttons": [
        {
          "text": "⭐️ 1",
          "data": "1"
        },
        {
          "text": "⭐️⭐️ 2",
          "data": "2"
        },
        {
          "text": "⭐️⭐️⭐️ 3",
          "data": "3"
        },
        {
          "text": "⭐️⭐️⭐️⭐️ 4",
          "data": "4"
        },
        {
          "text": "⭐️⭐️⭐️⭐️⭐️ 5",
          "data": "5"
        }
      ]
    },
    {
      "name": "screenshot",
      "label": "Upload a screenshot",
      "type": "file",
      "db_type": "TEXT",
      "required": false,
      "skippable": true,
      "description": "Please upload a screenshot if applicable.",
      "formatting": "Markdown",
      "buttons": []
    },
    {
      "name": "finish",
      "label": "Finish survey",
      "type": "text",
      "required": false,
      "skippable": false,
      "description": "You have completed the survey. Thank you for your time.",
      "formatting": "Markdown",
      "buttons": [
        {
          "text": "Review",
          "data": "review"
        }
      ]
    }
  ],
  "messages": {
    "submit": "🎉 Hooray! Your form has been submitted successfully! 🎊",
    "submit_button": "✅ Send it in!",
    "skip_button": "⏭️ Skip for now",
    "modify": "📝 Please enter a new value for %s:",
    "modify_button": "✏️ Change %s",
    "choose_option": "👇 Pick one of the options:",
    "review": "🔎 <b>Let's Review Your Inputs:</b>\n\n",
    "file_upload_success": "📂 Your file has been uploaded successfully!",
    "upload_another_button": "📁 Upload another file",
    "upload_another": "Need to upload another? Go ahead!",
    "finish_upload_button": "✔️ Done uploading",
    "finish_upload": "📤 Would you like to upload more files or finish?",
    "required_file": "⚠️ A file is needed for %s. Please upload one!",
    "required_select": "⚠️ You must make a selection! Choose one for %s from: %s.",
    "required_input": "⚠️ This %s is mandatory. Please enter a value.",
    "invalid_email": "🚨 That doesn’t look like a valid email. Try again!",
    "invalid_max_number": "⚠️ The value for %s must be at most %d. Please enter a valid number.",
    "invalid_min_number": "⚠️ The value for %s must be at least %d. Please enter a valid number.",
    "invalid_number": "⚠️ Oops! %s needs to be a number.",
    "invalid_format": "⚠️ The format for %s is incorrect. Please check and fix it.",
    "validation_error": "⚠️ Something went wrong with %s. Try again!",
    "invalid_max_length": "⚠️ %s is too long! Maximum %d characters allowed.",
    "invalid_min_length": "⚠️ %s is too short! Minimum %d characters required."
  }
}
```
### 📋 JSON Form Fields

* `form_name`: The name of the form.
* `table_name`: The name of the table in the database where the form data will be stored.
* `review_enabled`: A boolean indicating whether the form requires review before submission.
* `db`: The database connection string. (if db is disabled, this can be omitted)
* `submit_message`: The message to display after the form is submitted.
* `fields`: A list of fields in the form, where each field is defined by the `Field`.


### 📑 Field Definition
| Field Name | Description                                                                                                                                                        |
| --- |--------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `Name` | The name of the field.                                                                                                                                             |
| `Label` | The label displayed for the field.                                                                                                                                 |
| `Type` | The type of the field (option: 'text', 'number', 'email', 'select', 'file', 'photo','document','video'). ps. use 'file' if you want to store the user input files. |
| `DBType` | The database type of the field (e.g., "VARCHAR(255)", "TEXT", "INT", etc.).                                                                                        | |
| `Required` | A boolean indicating whether the field is required.                                                                                                                |
| `Skippable` | A boolean indicating whether the field can be skipped.                                                                                                             |
| `Description` | A brief description of the field.                                                                                                                                  |
| `Formatting` | The formatting of the field (e.g., "HTML", "Markdown").                                                                                                            |
| `Location` | The location of the file if you want to send `Type` "photo", "video", "document".                                                                                   | |
| `Buttons` | A list of buttons associated with the field.                                                                                                                       |
| `Options` | A list of options for the field (e.g., for select fields).                                                                                                         | |
| `Validation`| The validation rules for the field.                                                                                                                                |

## 🔍 Validation Fields

The Validation field is a sepecification that contains several fields that define the validation rules for the field. The following table explains each field:

| Field Name | Description |
| --- | --- |
| `MinLength` | The minimum length of the field. |
| `MaxLength` | The maximum length of the field. |
| `Regex` | A regular expression pattern to match the field value. |
| `Min` | The minimum value of the field (for numeric fields). |
| `Max` | The maximum value of the field (for numeric fields). |

These validation fields can be used to enforce various validation rules on the field, such as:

* Minimum and maximum length for text fields
* Regular expression pattern matching for text fields
* Minimum and maximum values for numeric fields


## 🗺️ Custom DB Type Mapping

The following table shows the custom mapping between database types and custom DB types:

| Custom DB Type   | MySQL          | PostgreSQL     | MongoDB   |
|------------------|----------------|----------------|-----------|
| `STRING`         | `VARCHAR(255)` | `VARCHAR(255)` | `string`  |
| `TEXT`           | `TEXT`         | `TEXT`         | `string`  |
| `NUMBER`         | `INT`          | `INTEGER`      | `int`     |
| `BOOLEAN`        | `BOOLEAN`      | `BOOLEAN`      | `bool`    |
| `DATETIME`       | `DATETIME`     | `TIMESTAMP`    | `date`    |
| `JSON`           | `JSON`         | `JSONB`        | `object`  |


## 💬 Messages (Optional)

The following table shows the Messages fields and usage. The placeholders are not mandatory but the exact amount and
placement will have the best effect. The placeholders can be `%s`, `%d` and `%v`. `%v` will be the most suitable.

| Field Name              | Description                                                                        | Default Value                                                           |
|-------------------------|------------------------------------------------------------------------------------|-------------------------------------------------------------------------|
| `submit`                | Show this message after users submit their form                                    | "🎉 Hooray! Your form has been submitted successfully! 🎊"              |
| `submit_button`         | Show `Submit` button message                                                       | : "✅ Send it in!"                                                       |                                         
| `skip_button`           | Show `Skip` button message when field has skippable set true                       | "⏭️ Skip for now"                                                       |                                          
| `modify`                | Show this message when users tried to modify an input                              | "📝 Please enter a new value for %s:"                                   |                      
| `modify_button`         | Show `Modify` button message when users is reviewing the inputs                    | "✏️ Change %s"                                                          |                                             
| `choose_option`         | Show this message when field type is `select`                                      | "👇 Pick one of the options:"                                           |                              
| `review`                | Show this message after users finished all input and `review_enable` is set `true` | "🔎 <b>Let's Review Your Inputs:</b>"                                   |                            
| `file_upload_success`   | Show this message when users successfully upload a file                            | "📂 Your file has been uploaded successfully!"                          |                                  
| `upload_another_button` | Show `Upload` Another button message                                               | "📁 Upload another file"                                                |                                                          
| `upload_another`        | Show this message after users uploaded a file                                      | "Need to upload another? Go ahead!"                                     |                                               
| `finish_upload_button`  | Show `Finish` Upload button message                                                | "✔️ Done uploading"                                                     |                                                       
| `finish_upload`         | Show this message when asking users to upload more or finish uploading.            | "📤 Would you like to upload more files or finish?"                     |                               
| `required_file`         | Show this message when users didn't upload any file and `required` is set `true`   | "⚠️ A file is needed for %s. Please upload one!"                        |                                  
| `required_select`       | Show this message when users didn't select an option and `required` is set `true`  | "⚠️ You must make a selection! Choose one for %s from: %s."             |                       
| `required_input`        | Show this message when users didn't add any input and `required` is set `true`     | "⚠️ This %s is mandatory. Please enter a value."                        |                                  
| `invalid_email`         | Show this message when users entered invalid email                                 | "🚨 That doesn’t look like a valid email. Try again!"                   |                             
| `invalid_max_number`    | Show this message when users entered a number over the maximum limit               | "⚠️ The value for %s must be at most %d. Please enter a valid number."  |            
| `invalid_min_number`    | Show this message when users entered a number under the minimum limit              | "⚠️ The value for %s must be at least %d. Please enter a valid number." |           
| `invalid_number`        | Show this message when users entered non-number value                              | "⚠️ Oops! %s needs to be a number."                                     |                                               
| `invalid_format`        | Show this message when users entered value is invalid/unsupported format           | "⚠️ The format for %s is incorrect. Please check and fix it."           |                     
| `validation_error`      | Show this message when users entered value failed the validation logic             | "⚠️ Something went wrong with %s. Try again!"                           |                                     
| `invalid_max_length`    | Show this message when users entered value over the maximum limit                  | "⚠️ %s is too long! Maximum %d characters allowed."                     |                               
| `invalid_min_length`    | Show this message when users entered value under the minimum limit                 | "⚠️ %s is too short! Minimum %d characters required."                   |                             


## 📂 Examples

For more information and examples, please see the [examples directory](/examples).

## 📜 License

This project is licensed under the MIT License.
MIT License