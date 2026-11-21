ai-expenses
===========

Utility for show daily expenses for OpenAI API and Google Gemini API.

  * OpenAI: uses [Costs API](https://platform.openai.com/docs/api-reference/usage/costs) with an admin key
  * Gemini: Gemini API has no usage/cost endpoint, so costs are read from [Cloud Billing export to BigQuery](https://cloud.google.com/billing/docs/how-to/export-data-bigquery), auth via [Application Default Credentials](https://cloud.google.com/docs/authentication/application-default-credentials)

Install
-------

    go install github.com/msoap/etc/ai-expenses@latest

Usage
-----

    ai-expenses [options]

    -provider string
        Provider: openai, gemini or both comma separated: openai,gemini (default "openai")
    -days int
        Number of days to retrieve (default 7)
    -total
        Show total cost
    -by-model
        Break out costs by model
    -verbose
        Show requests and tokens in addition to cost (when provider has them)
    -chart
        Show cost as horizontal bar chart at the end of lines
    -col
        Show each provider in its own columns (can't be used with -by-model)
    -html-report string
        Save HTML report to file, it always contains break out by models, requests and tokens (can't be used with -by-model, -verbose, -total, -chart, -col)
    -key string
        OpenAI API key (or set OPENAI_ADMIN_KEY env var)
    -gcp-project string
        GCP project with BigQuery billing export (default from GOOGLE_CLOUD_PROJECT env var or credentials)
    -bq-dataset string
        BigQuery dataset with billing export (default "billing_export")

Examples:

    $ OPENAI_ADMIN_KEY=sk-admin-... ai-expenses -days 3 -total
    2026-11-19 Th 0.1234
    2026-11-20 Fr 0.0567
    2026-11-21 Sa 0.0012
    Total: 0.1813

    $ ai-expenses -provider gemini -gcp-project my-gemini-project -days 30 -total

Output is one line per day (UTC): `YYYY-MM-DD Wd cost`, where `Wd` is two-letter weekday (`Mo`, `Tu`, `We`, `Th`, `Fr`, `Sa`, `Su`). Days without usage are shown as `0.0000`.

With `-by-model` each day (and `Total:`) is followed by its models sorted by cost, with `-verbose` requests and tokens columns are added:

    $ ai-expenses -by-model -verbose -days 2 -total
    date                        cost  requests  input_tokens  output_tokens
    2026-11-20 Fr             1.7500        12          1050            105
      gpt-4o-2024-08-06       1.5000        10          1000            100
      gpt-4o-mini-2024-07-18  0.2500         2            50              5
    2026-11-21 Sa             2.0000         3            70              7
      gpt-4o-mini-2024-07-18  2.0000         3            70              7
    Total:                    3.7500        15          1120            112
      gpt-4o-mini-2024-07-18  2.2500         5           120             12
      gpt-4o-2024-08-06       1.5000        10          1000            100

With `-chart` each line with cost ends with a bar up to 10 characters long, drawn with [tcg](https://github.com/msoap/tcg) in 2x3 mode (1px line in the middle of the character, so the resolution is half a character). Days and models of days are scaled to the most expensive day. The `Total:` line has no bar, its models are scaled to the total cost:

    $ ai-expenses -by-model -days 2 -total -chart
    2026-11-20 Fr             1.7500  🬋🬋🬋🬋🬋🬋🬋🬋🬋
      gpt-4o-2024-08-06       1.5000  🬋🬋🬋🬋🬋🬋🬋🬃
      gpt-4o-mini-2024-07-18  0.2500  🬋🬃
    2026-11-21 Sa             2.0000  🬋🬋🬋🬋🬋🬋🬋🬋🬋🬋
      gpt-4o-mini-2024-07-18  2.0000  🬋🬋🬋🬋🬋🬋🬋🬋🬋🬋
    Total:                    3.7500
      gpt-4o-mini-2024-07-18  2.2500  🬋🬋🬋🬋🬋🬋
      gpt-4o-2024-08-06       1.5000  🬋🬋🬋🬋

2x3 symbols are from Unicode "Symbols for Legacy Computing" block, they are supported by Iosevka font and some terminals.

### HTML report

    $ ai-expenses -provider openai,gemini -days 30 -html-report ai-expenses.html
    HTML report saved to ai-expenses.html

One self-contained HTML file (no external resources, template is embedded into the binary) with all providers together:

  * tiles: total cost, average per day, the most expensive day, requests, input/output tokens
  * cost share bar, daily cost chart (stacked columns), tables by provider/model and by day, heatmap cost by day and provider/model
  * "Break out by model" switch - shows all charts and tables by providers or by models

Charts, tables and the switch work without JavaScript (SVG + CSS), JavaScript only adds sorting tables by click on column header and remembers the switch state. Light and dark themes follow the system setting.

With `-provider openai,gemini` the providers are printed one after another, each under its own `openai:` / `gemini:` header. With `-col` they are printed side by side, one column per provider, works with `-total` and `-chart` (all bars are scaled to the most expensive day of all providers).

    $ ai-expenses -provider openai,gemini -days 2 -total -chart -col
    date          ┊ openai    ┊ gemini
    ┈┈┈┈┈┈┈┈┈┈┈┈┈┈┼┈┈┈┈┈┈┈┈┈┈┈┼┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈
    2026-11-20 Fr ┊ 1.7500 🬋🬃 ┊  0.0000
    2026-11-21 Sa ┊ 2.0000 🬋🬃 ┊ 12.5000 🬋🬋🬋🬋🬋🬋🬋🬋🬋🬋
    Total:        ┊ 3.7500    ┊ 12.5000

With `-verbose` each provider gets requests (`req`), input (`in`) and output (`out`) tokens columns after the cost (Gemini has no `req`):

    $ ai-expenses -provider openai,gemini -days 2 -total -verbose -col
    date          ┊ openai  req    in  out ┊ gemini    in  out
    ┈┈┈┈┈┈┈┈┈┈┈┈┈┈┼┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┼┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈
    2026-11-20 Fr ┊ 1.7500   12  1050  105 ┊  0.0000    0    0
    2026-11-21 Sa ┊ 2.0000    3    70    7 ┊ 12.5000  500   50
    Total:        ┊ 3.7500   15  1120  112 ┊ 12.5000  500   50

Where the data comes from:

  * OpenAI: costs per model from Costs API grouped by `line_item` (`"gpt-4o-2024-08-06, input"` → `gpt-4o-2024-08-06`), requests and tokens from [Usage API](https://platform.openai.com/docs/api-reference/usage) (completions, embeddings, moderations, images, audio) grouped by model. Model names in costs and in usage may differ for some line items, in that case they are shown as separate rows
  * Gemini: model and token direction are parsed from SKU description, e.g. `Generate content input token count gemini 2.5 flash lite short input text` → `gemini-2.5-flash-lite`, input tokens = `usage.amount` (`usage.unit` says `requests`, but it's a token count). Billing export has no request counts, so the `requests` column is not shown for Gemini. Non-token SKUs (e.g. grounding) are shown with SKU description as model name

OpenAI setup
------------

Create an admin key in [organization settings](https://platform.openai.com/settings/organization/admin-keys) (a regular project API key does not have access to Costs API) and pass it via `-key` or `OPENAI_ADMIN_KEY` env var.

Gemini setup
------------

The Gemini API key (`AIza...`) can't be used: it only allows calling models. Costs are taken from billing data, so the setup is:

### 1. Paid tier

The Gemini project must be linked to a billing account (paid tier). Free tier usage is not billed and is not in billing export.

### 2. BigQuery dataset

Create a dataset in some project, for example in the Gemini project. Use multi-region location (`US` or `EU`) and don't set default table expiration:

    gcloud services enable bigquery.googleapis.com bigquerydatatransfer.googleapis.com --project=PROJECT_ID
    bq mk --dataset --location=US PROJECT_ID:billing_export

### 3. Billing export

Can be enabled only in the console: Billing → select billing account → Billing export → BigQuery export → **Standard usage cost** → Edit settings → select project and `billing_export` dataset → Save.

Google creates table `gcp_billing_export_v1_<BILLING_ACCOUNT_ID>` itself. The first data appears in a few hours (up to a day), initial backfill of history can take a day or two. After that data is updated several times a day, costs for the last 1-2 days can still change.

Check that data is there and which services are exported:

    bq query --use_legacy_sql=false \
      'SELECT service.description, DATE(usage_start_time) AS day, SUM(cost) AS cost
       FROM `PROJECT_ID.billing_export.gcp_billing_export_v1_*`
       GROUP BY 1, 2 ORDER BY day DESC LIMIT 50'

Gemini costs are in rows with `service.description = 'Gemini API'`. The utility shows net cost (after credits) summed for all projects in the billing account.

### 4. Credentials

The account that runs the utility needs these roles in the project with the dataset:

  * `roles/bigquery.jobUser` - for run queries
  * `roles/bigquery.dataViewer` - for read billing export table

#### Locally

    gcloud auth application-default login
    gcloud auth application-default set-quota-project PROJECT_ID
    ai-expenses -provider gemini -gcp-project PROJECT_ID

Or use a service account key: `GOOGLE_APPLICATION_CREDENTIALS=/path/to/key.json`.

#### On GCP VM

No keys and no `gcloud` on the VM are needed, the token is taken from the metadata server for the VM service account. Run these commands from your machine or Cloud Shell.

Check the VM service account and its access scopes, scopes must include `https://www.googleapis.com/auth/cloud-platform` (or BigQuery):

    gcloud compute instances describe VM_NAME --zone=ZONE \
      --format='value(serviceAccounts[].email,serviceAccounts[].scopes)'

If scopes are not enough, change them (VM must be stopped):

    gcloud compute instances stop VM_NAME --zone=ZONE
    gcloud compute instances set-service-account VM_NAME --zone=ZONE \
      --service-account=SA_EMAIL --scopes=cloud-platform
    gcloud compute instances start VM_NAME --zone=ZONE

Grant roles to the VM service account in the project with the dataset (it can be a different project than the VM):

    gcloud projects add-iam-policy-binding PROJECT_ID \
      --member=serviceAccount:SA_EMAIL --role=roles/bigquery.jobUser
    gcloud projects add-iam-policy-binding PROJECT_ID \
      --member=serviceAccount:SA_EMAIL --role=roles/bigquery.dataViewer

Build for the VM and run it there:

    GOOS=linux GOARCH=amd64 go build -o ai-expenses .
    ./ai-expenses -provider gemini -gcp-project PROJECT_ID

If the dataset is in the same project as the VM, `-gcp-project` can be omitted.

Zone and VM name can be found from inside the VM:

    curl -s -H "Metadata-Flavor: Google" http://metadata.google.internal/computeMetadata/v1/instance/zone
    curl -s -H "Metadata-Flavor: Google" http://metadata.google.internal/computeMetadata/v1/instance/name

### Cost

Billing export is free. BigQuery free tier is 10 GB storage and 1 TiB queries per month, billing data of a personal account is a few MB, so for this utility it's practically $0.

### Troubleshooting

  * `403 ... insufficient authentication scopes` - VM access scopes don't include BigQuery, see above
  * `403 ... Access Denied` / `permission bigquery.jobs.create` - roles are not granted in the dataset project, or wrong `-gcp-project`
  * `Not found: Table ...` - billing export has not written any data yet
  * all days are `0.0000` - check with the query above that the table already has data for these days (backfill may still be in progress) and that there are `Gemini API` rows
