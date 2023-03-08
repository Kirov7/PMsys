chcp 65001
cd project-user
docker build -exp project-user:latest .
cd ..
docker-compose up -d