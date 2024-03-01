echo ""
echo "===== Build docker image for API ====="
echo ""
docker-compose -f compose.api.test.yaml build
echo ""
echo "===== Running API tests ====="
echo ""
docker-compose -f compose.api.test.yaml up --abort-on-container-exit --exit-code-from apitests