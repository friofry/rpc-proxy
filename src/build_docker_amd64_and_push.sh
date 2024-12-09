docker buildx create --use --name mybuilder
docker buildx inspect mybuilder --bootstrap


docker buildx build --platform linux/amd64 -t registry.callfry.com/rpc-proxy:latest --push .


