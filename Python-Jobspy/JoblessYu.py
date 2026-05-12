import os
from jobspy import scrape_jobs
from dotenv import load_dotenv
import pandas as pd

load_dotenv()

try:
    import psycopg
except ImportError:
    psycopg = None


def _to_nullable(value):
    return None if pd.isna(value) else value


def save_jobs_to_neon(jobs_df):
    database_url = os.getenv("DATABASE_URL")
    if not database_url:
        print("DATABASE_URL not set. Skipping Neon save.")
        return

    if psycopg is None:
        print("psycopg is not installed. Run: pip install psycopg[binary]")
        return

    create_table_sql = """
    CREATE TABLE IF NOT EXISTS jobs (
        id BIGSERIAL PRIMARY KEY,
        job_id TEXT,
        site TEXT NOT NULL,
        job_url TEXT NOT NULL,
        title TEXT,
        company TEXT,
        location TEXT,
        job_type TEXT,
        fetched_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        UNIQUE (site, job_url)
    );
    """

    insert_sql = """
    INSERT INTO jobs (job_id, site, job_url, title, company, location, job_type)
    VALUES (%s, %s, %s, %s, %s, %s, %s)
    ON CONFLICT (site, job_url)
    DO UPDATE SET
        job_id = EXCLUDED.job_id,
        title = EXCLUDED.title,
        company = EXCLUDED.company,
        location = EXCLUDED.location,
        job_type = EXCLUDED.job_type,
        fetched_at = NOW();
    """

    rows = [
        (
            _to_nullable(row["id"]),
            row["site"],
            row["job_url"],
            _to_nullable(row["title"]),
            _to_nullable(row["company"]),
            _to_nullable(row["location"]),
            _to_nullable(row["job_type"]),
        )
        for _, row in jobs_df.iterrows()
    ]

    with psycopg.connect(database_url) as conn:
        with conn.cursor() as cur:
            cur.execute(create_table_sql)
            cur.executemany(insert_sql, rows)
        conn.commit()

    print(f"{len(rows)} jobs upserted to Neon.")

def JobScan():
    print("JoblessYu is looking for jobs =w=")
    
    # Scrape jobs using JobSpy.
    jobs = scrape_jobs(
        site_name=["indeed", "linkedin"],
        search_term="IT Support",
        location="vietnam",
        results_wanted=20,
        hours_old=24*7,
        country_indeed='vietnam',
    )

    if jobs.empty:
        print("There are no jobs at the moment :c")
        return
    
    #Pandas DataFrame to JSON
    available_filters = ["id", "site", "job_url", "title", "company", "location", "job_type"]
    jobs = jobs[available_filters]
    jobs.to_json("jobs.json", orient="records", indent=4, force_ascii=False)
    print(f"{len(jobs)} jobs saved to jobs.json =w=")
    save_jobs_to_neon(jobs)
    

if __name__ == "__main__":
    JobScan()