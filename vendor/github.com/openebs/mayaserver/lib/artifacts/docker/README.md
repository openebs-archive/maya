## BUILD & RUN m-apiserver's image

```bash
# build & tag the image
sudo docker build -t openebs/m-apiserver:latest -t openebs/m-apiserver:0.2-RC4 .

# run the image as a docker container
sudo docker run -itd openebs/m-apiserver:latest

# verify the container
sudo docker ps
sudo docker ps -a
sudo docker logs <Container-ID>
sudo docker inspect <Container-ID>

# verify the m-apiserver service within this container
curl http://<Container-IP>:5656/latest/meta-data/instance-id

# Verify INI file inside the container
amit:docker$ sudo docker exec -it <Container-ID> bash
root@921f974ee490:/# 
root@921f974ee490:/# cat /etc/mayaserver/orchprovider/nomad_global.INI
```

## To run the image with custom ENV values 

Refer [docker-compose](../docker-compose/README.md)

## Cleaning up Docker Images & Containers

> This might be required in your dev/test machine or your own laptop.

```bash
# Stop all Containers
sudo docker stop $(sudo docker ps -a -q)

# Remove all Containers & its Associated Volumes
sudo docker rm -v $(sudo docker ps -a -q)

# Remove all Images
sudo docker rmi $(sudo docker images -q)
```

## TODO

- Check if ENV variables get overridden if specified in `docker run` command
- Make maya api server's port configurable
- Add sample K8s pod spec that deploys m-apiserver as a K8s pod.
- Run m-apiserver up command without -bind option
- Follow dockerfile best practices
- Follow entrypoint best practices
- Follow scripting best practices
- Run linting on dockerfile
- Run linting on script
- Use a GoLang binary script than an entrypoint shell script
