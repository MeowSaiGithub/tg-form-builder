# 📝 Dynamic Form Telegram Bot

A powerful Telegram bot that dynamically generates forms based on a provided JSON configuration. It supports storing responses in MySQL, PostgreSQL, or MongoDB and can send data via a webhook.

## ✨ Features

- 📜 **Dynamic Form Generation** – Define forms via JSON.
- 💾 **Database Support** – Store responses in **MySQL, PostgreSQL, or MongoDB**.
- 🔄 **Webhook Integration** – Send data to external services.
- 📸 **Media Support** – Forms can include **photos, videos, and documents**.
- ✅ **Validation & Preprocessing** – Supports input validation and required fields.
- 🎨 **Customizable Buttons** – Forms can include inline buttons for user interaction.
- 🛠️ **Debug Mode & Memory Load** – Helps with performance tuning and debugging.
- 🔧 **Custom Executables** – Custom Executables.


## 🚀 Getting Start

Download the bot from [Latest Release](https://github.com/telesign/tgform/releases/latest). You can download with multiple options -

- `all` tag include all databases `mysql`, `postgres`, and `mongo`
- `mongo` tag include only `mongo` database
- `postgres` tag include only `postgres` database
- `mysql` tag include only `mysql` database
- `none` tag include only `blank` database


### 🛠️ Commands
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
  use_adaptor: "mongo"  # Choose from mysql, postgres, or mongo
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
  ]
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

| Custom DB Type | MySQL | PostgreSQL | MongoDB |
| --- | --- | --- | --- |
| `STRING` | `VARCHAR(255)` | `VARCHAR(255)` | `string` |
| `TEXT` | `TEXT` | `TEXT` | `string` |
| `NUMBER` | `INT` | `INTEGER` | `int` |
| `BOOLEAN` | `BOOLEAN` | `BOOLEAN` | `bool` |
| `DATETIME` | `DATETIME` | `TIMESTAMP` | `date` |
| `JSON` | `JSON` | `JSONB` | `object` |

## 📂 Examples

For more information and examples, please see the [examples directory](/examples).

## 📜 License

This project is licensed under the MIT License.
MIT License