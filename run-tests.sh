echo ""
echo "===== Build docker images ====="
echo ""
docker-compose -f compose.web.test.yaml build
echo ""
echo "===== Running webserver tests ====="
echo ""
docker-compose -f compose.web.test.yaml up --abort-on-container-exit --exit-code-from webtests