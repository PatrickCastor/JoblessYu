import json
from jobspy import scrape_jobs
import pandas as pd

def JobScan():
    print("JoblessYu is looking for jobs :3")
    
    # Scrape jobs using JobSpy.
    jobs = scrape_jobs(
        site_name=["indeed", "linkedin"],
        search_term="IT Support",
        location="Vietnam",
        results_wanted=30,
        hours_old=247,
        country_indeed='vietnam',
    )

    if jobs.empty:
        print("There are no jobs at the moment :c")
        return

    # Pandas DataFrame to JSON.
    

if __name__ == "__main__":
    JobScan()