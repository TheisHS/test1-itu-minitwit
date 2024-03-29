# Container image that runs your code
FROM python:3

RUN pip install --no-cache-dir pytest==8.1.1 \
 && pip install --no-cache-dir requests==2.31.0

WORKDIR /
COPY ./src-tests ./
COPY ../minitwit.db .
COPY ../schema.sql .

CMD ["pytest", "refactored_tests.py"]