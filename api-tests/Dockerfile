# Container image that runs your code
FROM python:3

RUN pip install pytest
RUN pip install requests

COPY ./api-tests ./
COPY ../minitwit.db .
COPY ../schema.sql .

RUN go mod download && go mod verify

CMD ["pytest", "api_sim_test.py"]