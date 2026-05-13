# JoblessYu
Discordgo - Job Announcer Bot

## Get Started with Neon

This repo can now save scraped jobs to Neon using `DATABASE_URL`.

### 1) Set your Neon connection string

Your local `.env` is already gitignored. Add:


### 2) Install Python dependencies

From the project root:

```bash
source .venv/bin/activate
pip install pandas python-jobspy "psycopg[binary]"
```

### 3) Run the scraper

```bash
cd Python-Jobspy
python JoblessYu.py
```

### 4) Verify jobs in Neon

The script creates a `jobs` table automatically and upserts records by `(site, job_url)`.

Example SQL:

```sql
SELECT site, title, company, location, fetched_at
FROM jobs
ORDER BY fetched_at DESC
LIMIT 20;
```
