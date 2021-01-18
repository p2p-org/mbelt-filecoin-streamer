docker-compose up -d --build db redis email redash-server redash-scheduler redash-worker
echo "Waiting a bit for redash containers to start"
/bin/sleep 15
docker-compose run --rm redash-server create_db
FILE=./.env
if [ -f "$FILE" ]; then
    echo ".env file exists."
else
    echo ".env file does not exist. Please create .env similar to example.env"
    exit 1
fi
python3 dashboard.py