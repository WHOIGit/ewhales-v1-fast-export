# eWHALES v1 Monorepo

Welcome to the eWHALES v1 Monorepo. This overarching repository contains all the project work related to the eWHALES v1 dataset, which involves the extraction, processing, and analysis of digitized 19th-century New England whaling logbooks.

This repository is organized into distinct sub-repositories, each focusing on a specific part of the data lifecycle. 

## 📁 Repository Structure

### 1. Data Pipeline (`/v1-data-pipeline`)
**Purpose:** Cleaning, gap-filling, and analyzing the whaling logbook data.

The `v1-data-pipeline` sub-repository contains Jupyter notebooks and Python scripts that transform raw database exports into a validated, tiered scientific dataset. 

**Key Features:**
- Cleans and standardizes raw text entries (e.g., coordinates, wind force mapped to the Beaufort scale).
- Infills missing coordinates across gaps of varying sizes to produce tiered datasets (Tiers 1-4).
- Generates standard publication figures, exploratory visualizations, and per-logbook metadata.
- Cross-references ship sightings between different logbooks to evaluate data consistency.

[Read more in the Data Pipeline README](./v1-data-pipeline/README.md)

### 2. Fast Export Tools (`/v1-fast-export`)
**Purpose:** Fast, scalable extraction of raw data from the live eWHALES v1 database.

The `v1-fast-export` sub-repository provides Go-based tools and Python scripts to efficiently export records from the database into the exact CSV format expected by the downstream data pipeline.

**Key Features:**
- A robust Go-based gRPC client/server architecture for live, streaming database exports.
- A standalone local exporter tool for direct database connections.
- Python tools (`convert_inserts.py`, `anonymize.py`) for processing and anonymizing raw SQL database backup dumps.
- Docker support for running the export tools seamlessly without local dependencies.

[Read more in the Fast Export README](./v1-fast-export/README.md)

---

## Getting Started

To get started with either project, navigate to the respective directory and follow the instructions provided in their individual `README.md` files.

```bash
# To work on the data processing and analysis pipeline:
cd v1-data-pipeline

# To work on the database extraction tools:
cd v1-fast-export
```
