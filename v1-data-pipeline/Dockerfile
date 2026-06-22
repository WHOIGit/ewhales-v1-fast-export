FROM condaforge/miniforge3:latest

# Set working directory
WORKDIR /app

# Copy environment file
COPY conda-environment.yml /app/

# Install the environment using mamba for faster resolution
RUN mamba env create -f conda-environment.yml && mamba clean -afy

# Copy the rest of the repo
COPY . /app/

# Expose Jupyter port
EXPOSE 8888

# Command to run jupyter notebook inside the conda environment
CMD ["conda", "run", "--no-capture-output", "-n", "ewhales", "jupyter", "notebook", "--ip=0.0.0.0", "--port=8888", "--no-browser", "--allow-root", "--NotebookApp.token='ewhales'"]
