# Container image that runs your code
FROM python:3

RUN pip install pytest
RUN pip install requests

COPY ./src-tests ./
COPY ../minitwit.db .
COPY ../schema.sql .

CMD ["pytest", "refactored_tests.py"]