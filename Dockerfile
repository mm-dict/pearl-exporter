FROM python:3.13-trixie

WORKDIR /sanic

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY pearl_exporter.py prober.py bao.py ./

EXPOSE 8000

ENTRYPOINT ["sanic", "pearl_exporter:app", "--port=8000", "--host=0.0.0.0", "--workers=4"]
